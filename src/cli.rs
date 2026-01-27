use clap::builder::styling::{AnsiColor, Color, Style};
use clap::{ArgAction, Parser};
use std::path::PathBuf;

#[derive(Parser, Debug)]
#[command(
    name = "repo-pack",
    version,
    about = "Download files from GitHub repository directories",
    long_about = "Repo-Pack is a tool designed to download files from a specified GitHub repository directory, preserving the directory structure"
)]
#[command(styles = get_styles())]
pub struct Cli {
    /// GitHub repository URL
    ///
    /// Example: https://github.com/owner/repo/tree/main/path/to/dir
    #[arg(value_name = "URL")]
    pub url: String,

    /// GitHub personal access token
    ///
    /// Can also be set via GITHUB_TOKEN environment variable.
    /// Required for private repositories.
    #[arg(long, env = "GITHUB_TOKEN", value_name = "TOKEN")]
    pub token: Option<String>,

    /// Output directory for downloaded files
    #[arg(long, short = 'o', default_value = ".", value_name = "DIR")]
    pub output: PathBuf,

    /// Concurrent download limit
    ///
    /// Maximum number of files to download simultaneously.
    /// Higher values may improve speed but could trigger rate limits.
    #[arg(long, short = 'l', default_value = "5", value_name = "NUM", value_parser = clap::value_parser!(u64).range(1..))]
    pub limit: u64,

    /// Preview files without downloading
    ///
    /// Shows what would be downloaded without actually downloading files.
    #[arg(long, short = 'n')]
    pub dry_run: bool,

    /// Skip files that already exist locally
    ///
    /// Useful for resuming interrupted downloads.
    #[arg(long, short = 'r')]
    pub resume: bool,

    /// Use verbose output
    ///
    /// Use multiple times for more verbosity (e.g., -vv)
    #[arg(long, short, action = ArgAction::Count, conflicts_with = "quiet")]
    pub verbose: u8,

    /// Use quiet output
    ///
    /// Use multiple times for less output (e.g., -qq for silent)
    #[arg(long, short, action = ArgAction::Count, conflicts_with = "verbose")]
    pub quiet: u8,

    /// Disable progress bar output
    #[arg(long)]
    pub no_progress: bool,
}

fn get_styles() -> clap::builder::Styles {
    clap::builder::Styles::styled()
        .usage(
            Style::new()
                .bold()
                .underline()
                .fg_color(Some(Color::Ansi(AnsiColor::Yellow))),
        )
        .header(
            Style::new()
                .bold()
                .underline()
                .fg_color(Some(Color::Ansi(AnsiColor::Yellow))),
        )
        .literal(Style::new().fg_color(Some(Color::Ansi(AnsiColor::Green))))
        .invalid(
            Style::new()
                .bold()
                .fg_color(Some(Color::Ansi(AnsiColor::Red))),
        )
        .error(
            Style::new()
                .bold()
                .fg_color(Some(Color::Ansi(AnsiColor::Red))),
        )
        .valid(
            Style::new()
                .bold()
                .underline()
                .fg_color(Some(Color::Ansi(AnsiColor::Green))),
        )
        .placeholder(Style::new().fg_color(Some(Color::Ansi(AnsiColor::White))))
}
