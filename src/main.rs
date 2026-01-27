use anstream::eprintln;
use clap::Parser;
use miette::Result;
use owo_colors::OwoColorize;
use repo_pack::{Cli, Config};

#[tokio::main(flavor = "current_thread")]
async fn main() -> Result<()> {
    let mut cli = Cli::parse();

    let config = Config::load()?;

    if cli.token.is_none() {
        if let Some(token) = config.read_token() {
            cli.token = Some(token);
        }
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
