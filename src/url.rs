use crate::error::RepoPackError;

/// A parsed GitHub repository URL with extracted components.
///
/// Supports two URL formats:
/// - Explicit branch: `https://github.com/{owner}/{repo}/tree/{branch}/{path}`
/// - Default branch: `https://github.com/{owner}/{repo}/{path}`
#[derive(Debug, Clone)]
pub struct ParsedUrl {
    pub owner: String,
    pub repo: String,
    /// Branch/tag/commit. None means default branch needs to be fetched.
    pub git_ref: Option<String>,
    pub dir: String,
}

impl ParsedUrl {
    /// Parses a GitHub URL into its components.
    ///
    /// Accepts URLs with or without explicit branch specification.
    /// Returns an error if the URL uses `/blob/` (single file) instead of `/tree/` (directory).
    pub fn parse(url_str: &str) -> Result<Self, RepoPackError> {
        let url = url::Url::parse(url_str).map_err(|_| RepoPackError::InvalidUrl {
            url: url_str.to_string(),
            hint: "Expected format: https://github.com/owner/repo[/tree/branch][/path]".to_string(),
        })?;

        let path = url.path();

        if path.contains("/blob/") {
            return Err(RepoPackError::InvalidUrl {
                url: url_str.to_string(),
                hint: "Use '/tree/' for directories, not '/blob/' (which is for single files)"
                    .to_string(),
            });
        }

        let parts: Vec<&str> = path.split('/').filter(|s| !s.is_empty()).collect();

        if parts.len() < 2 {
            return Err(RepoPackError::InvalidUrl {
                url: url_str.to_string(),
                hint: "URL must include owner and repo: https://github.com/owner/repo".to_string(),
            });
        }

        let owner = parts[0].to_string();
        let repo = parts[1].to_string();

        if parts.len() > 2 && parts[2] == "tree" {
            if parts.len() < 4 {
                return Err(RepoPackError::InvalidUrl {
                    url: url_str.to_string(),
                    hint: "Missing branch after /tree/".to_string(),
                });
            }
            let git_ref = parts[3].to_string();
            let dir = if parts.len() > 4 {
                parts[4..].join("/")
            } else {
                String::new()
            };
            return Ok(Self {
                owner,
                repo,
                git_ref: Some(git_ref),
                dir,
            });
        }

        let dir = if parts.len() > 2 {
            parts[2..].join("/")
        } else {
            String::new()
        };

        Ok(Self {
            owner,
            repo,
            git_ref: None,
            dir,
        })
    }

    /// Returns true if the URL didn't specify a branch and needs to fetch the default.
    pub fn needs_default_branch(&self) -> bool {
        self.git_ref.is_none()
    }

    /// Sets the git ref (branch/tag/commit).
    pub fn set_git_ref(&mut self, git_ref: String) {
        self.git_ref = Some(git_ref);
    }

    /// Returns the git ref, falling back to "main" if not set.
    pub fn git_ref(&self) -> &str {
        self.git_ref.as_deref().unwrap_or("main")
    }
}
