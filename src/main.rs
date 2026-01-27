use anstream::eprintln;
use clap::Parser;
use miette::Result;
use owo_colors::OwoColorize;
use repo_pack::Cli;

#[tokio::main(flavor = "current_thread")]
async fn main() -> Result<()> {
    let cli = Cli::parse();

    if cli.limit > 100 {
        eprintln!(
            "{}: high concurrent download limit ({}) may cause rate limiting",
            "warning".yellow().bold(),
            cli.limit
        );
    }

    // TODO: implement download logic

    Ok(())
}
