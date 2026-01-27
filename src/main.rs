use anstream::{eprintln, println};
use clap::Parser;
use miette::Result;
use owo_colors::OwoColorize;
use repo_pack::{
    CancellationToken, Cli, Config, DownloadOptions, DownloadProgress, GitHubProvider, ParsedUrl,
    download_files,
};
use std::sync::Arc;
use std::sync::atomic::{AtomicBool, Ordering};
use std::time::Instant;

#[tokio::main(flavor = "current_thread")]
async fn main() -> Result<()> {
    let mut cli = Cli::parse();

    let config = Config::load()?;

    if cli.token.is_none()
        && let Some(token) = config.read_token()
    {
        cli.token = Some(token);
    }

    if cli.limit > 100 {
        eprintln!(
            "{}: high concurrent download limit ({}) may cause rate limiting",
            "warning".yellow().bold(),
            cli.limit
        );
    }

    let mut parsed_url = ParsedUrl::parse(&cli.url)?;
    let provider = GitHubProvider::new()?;

    let files = provider
        .list_files(&mut parsed_url, cli.token.as_deref())
        .await?;

    if files.is_empty() {
        println!("No files found in {}", cli.url.cyan());
        return Ok(());
    }

    if cli.dry_run {
        println!(
            "Dry run — {} file(s) ready to download",
            files.len().to_string().cyan()
        );
        if cli.verbose > 0 {
            for (i, file) in files.iter().enumerate() {
                println!("  {}. {}", i + 1, file.dimmed());
            }
        }
        return Ok(());
    }

    let base_dir = std::path::Path::new(&parsed_url.dir)
        .file_name()
        .and_then(|s| s.to_str())
        .unwrap_or(&parsed_url.dir);

    let total_files = files.len() as u64;
    let progress = DownloadProgress::new(total_files, cli.quiet > 0 || cli.no_progress);

    let cancelled: CancellationToken = Arc::new(AtomicBool::new(false));
    let cancelled_handler = cancelled.clone();

    ctrlc::set_handler(move || {
        cancelled_handler.store(true, Ordering::SeqCst);
    })
    .expect("failed to set Ctrl-C handler");

    let options = DownloadOptions {
        base_dir,
        output_dir: &cli.output,
        concurrency_limit: cli.limit as usize,
        resume: cli.resume,
        verbose: cli.verbose > 0,
    };

    let start = Instant::now();
    let result = download_files(
        &provider,
        &parsed_url,
        files,
        options,
        &progress,
        &cancelled,
    )
    .await;
    let duration = start.elapsed();

    if result.cancelled {
        let incomplete = total_files - result.downloaded - result.skipped;
        eprintln!(
            "\n{}: download cancelled with {} incomplete file(s)",
            "cancelled".yellow().bold(),
            incomplete
        );
        std::process::exit(1);
    }

    print_summary(&result, total_files, duration, cli.quiet > 0);

    if !result.errors.is_empty() && cli.verbose > 0 {
        eprintln!();
        for (path, err) in &result.errors {
            eprintln!("  {}: {} — {}", "failed".red(), path, err);
        }
    }

    Ok(())
}

fn print_summary(
    result: &repo_pack::DownloadResult,
    total: u64,
    duration: std::time::Duration,
    quiet: bool,
) {
    if quiet {
        return;
    }

    let mut parts = vec![format!(
        "{}/{}",
        result.downloaded.to_string().green(),
        total
    )];

    parts.push(" downloaded".to_string());

    if result.skipped > 0 {
        parts.push(format!(", {} skipped", result.skipped.to_string().yellow()));
    }

    if result.failed > 0 {
        parts.push(format!(", {} failed", result.failed.to_string().red()));
    }

    parts.push(
        format!(" [{:.3}s]", duration.as_secs_f64())
            .dimmed()
            .to_string(),
    );

    println!("{}", parts.join(""));
}

#[cfg(test)]
mod tests {
    use repo_pack::ParsedUrl;

    #[test]
    fn test_parse_valid_url() {
        let url = "https://github.com/owner/repo/tree/main/path/to/dir";
        let parsed = ParsedUrl::parse(url).unwrap();
        assert_eq!(parsed.owner, "owner");
        assert_eq!(parsed.repo, "repo");
        assert_eq!(parsed.git_ref, "main");
        assert_eq!(parsed.dir, "path/to/dir");
    }

    #[test]
    fn test_parse_blob_url_fails() {
        let url = "https://github.com/owner/repo/blob/main/file.rs";
        let result = ParsedUrl::parse(url);
        assert!(result.is_err());
    }
}
