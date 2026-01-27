use anstream::eprintln;
use clap::Parser;
use miette::Result;
use owo_colors::OwoColorize;
use repo_pack::{Cli, Config};

#[tokio::main(flavor = "current_thread")]
async fn main() -> Result<()> {
    let mut cli = Cli::parse();

    let config = Config::load()?;

    if cli.token.is_none() && let Some(token) = config.read_token() {
        cli.token = Some(token);
    }

    if cli.limit > 100 {
        eprintln!(
            "{}: high concurrent download limit ({}) may cause rate limiting",
            "warning".yellow().bold(),
            cli.limit
        );
    }

    Ok(())
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
