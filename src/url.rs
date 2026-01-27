use crate::error::RepoPackError;
use regex::Regex;
use std::sync::LazyLock;

static URL_REGEX: LazyLock<Regex> = LazyLock::new(|| {
    Regex::new(r"^/([^/]+)/([^/]+)/tree/([^/]+)/(.*)").expect("URL regex is valid")
});

/// A parsed GitHub repository URL with extracted components.
///
/// Represents the structure: `https://github.com/{owner}/{repo}/tree/{git_ref}/{dir}`
#[derive(Debug, Clone)]
pub struct ParsedUrl {
    pub owner: String,
    pub repo: String,
    pub git_ref: String,
    pub dir: String,
}

impl ParsedUrl {
    /// Parses a GitHub URL into its components.
    ///
    /// Accepts URLs in the format `https://github.com/owner/repo/tree/branch/path`.
    /// Returns an error if the URL uses `/blob/` (single file) instead of `/tree/` (directory).
    pub fn parse(url_str: &str) -> Result<Self, RepoPackError> {
        let url = url::Url::parse(url_str).map_err(|_| RepoPackError::InvalidUrl {
            url: url_str.to_string(),
            hint: "Expected format: https://github.com/owner/repo/tree/branch/path".to_string(),
        })?;

        let path = url.path();

        if path.contains("/blob/") {
            return Err(RepoPackError::InvalidUrl {
                url: url_str.to_string(),
                hint: "Make sure the URL contains '/tree/' (not '/blob/')".to_string(),
            });
        }

        let captures = URL_REGEX
            .captures(path)
            .ok_or_else(|| RepoPackError::InvalidUrl {
                url: url_str.to_string(),
                hint: "Expected format: https://github.com/owner/repo/tree/branch/path".to_string(),
            })?;

        Ok(Self {
            owner: captures[1].to_string(),
            repo: captures[2].to_string(),
            git_ref: captures[3].to_string(),
            dir: captures[4].to_string(),
        })
    }
}
