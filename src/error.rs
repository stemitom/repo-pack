use miette::Diagnostic;
use std::path::PathBuf;
use thiserror::Error;

#[derive(Error, Debug, Diagnostic)]
pub enum RepoPackError {
    #[error("invalid url: {url}")]
    #[diagnostic(help("{hint}"))]
    InvalidUrl { url: String, hint: String },

    #[error("rate limit exceeded, resets at {reset_time}")]
    #[diagnostic(help("Try using a personal access token with --token"))]
    RateLimited { reset_time: String },

    #[error("repository not found: {owner}/{repo}")]
    #[diagnostic(help("{hint}"))]
    NotFound {
        owner: String,
        repo: String,
        hint: String,
    },

    #[error("authentication required for private repository")]
    #[diagnostic(help("Use --token to provide a GitHub personal access token"))]
    AuthRequired,

    #[error("failed to download {path}")]
    DownloadFailed {
        path: String,
        #[source]
        source: reqwest::Error,
    },

    #[error("failed to save file {}", path.display())]
    IoError {
        path: PathBuf,
        #[source]
        source: std::io::Error,
    },

    #[error("path traversal detected: {path}")]
    #[diagnostic(help("The path attempts to escape the output directory"))]
    PathTraversal { path: String },

    #[error("failed to load config")]
    ConfigLoad {
        #[source]
        source: std::io::Error,
    },

    #[error("failed to parse config")]
    ConfigParse {
        #[source]
        source: serde_json::Error,
    },

    #[error("failed to save config")]
    ConfigSave {
        #[source]
        source: std::io::Error,
    },
}
