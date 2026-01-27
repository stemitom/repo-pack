use crate::error::RepoPackError;
use crate::url::ParsedUrl;
use bytes::Bytes;
use reqwest::{Client, StatusCode};
use serde::Deserialize;
use std::time::Duration;

#[derive(Debug, Deserialize)]
struct TreeItem {
    #[serde(rename = "type")]
    item_type: String,
    path: String,
}

#[derive(Debug, Deserialize)]
struct TreeResponse {
    tree: Vec<TreeItem>,
    truncated: bool,
}

#[derive(Debug, Deserialize)]
struct ContentItem {
    #[serde(rename = "type")]
    item_type: String,
    path: String,
}

/// GitHub API client for listing and downloading repository contents.
///
/// Handles authentication, rate limiting, and Git LFS transparently.
pub struct GitHubProvider {
    client: Client,
}

impl GitHubProvider {
    /// Creates a new GitHub provider with a configured HTTP client.
    pub fn new() -> Result<Self, RepoPackError> {
        let client = Client::builder()
            .timeout(Duration::from_secs(30))
            .pool_max_idle_per_host(10)
            .pool_idle_timeout(Duration::from_secs(90))
            .user_agent("repo-pack/0.1.0")
            .build()
            .map_err(|e| RepoPackError::DownloadFailed {
                path: "client initialization".to_string(),
                source: e,
            })?;

        Ok(Self { client })
    }

    fn download_err(&self, path: &str, source: reqwest::Error) -> RepoPackError {
        RepoPackError::DownloadFailed {
            path: path.to_string(),
            source,
        }
    }

    /// List files in a GitHub repository directory.
    ///
    /// Handles branches with slashes (e.g., `feature/my-branch`) by iteratively
    /// adjusting the ref until the trees API succeeds. Falls back to contents API
    /// when the tree response is truncated (large repositories).
    pub async fn list_files(
        &self,
        parsed_url: &mut ParsedUrl,
        token: Option<&str>,
    ) -> Result<Vec<String>, RepoPackError> {
        let decoded_dir = urlencoding::decode(&parsed_url.dir)
            .map_err(|_| RepoPackError::InvalidUrl {
                url: parsed_url.dir.clone(),
                hint: "Failed to decode directory path".to_string(),
            })?
            .into_owned();

        let mut dir_parts: Vec<String> = decoded_dir
            .split('/')
            .filter(|s| !s.is_empty())
            .map(String::from)
            .collect();

        let mut files = Vec::new();
        let mut truncated = false;

        // Branch-with-slash resolution: if URL is /owner/repo/tree/feature/branch/src/lib
        // and branch is actually "feature/branch", the initial parse gives:
        //   ref="feature", dir="branch/src/lib"
        // On 404, we shift: ref="feature/branch", dir="src/lib"
        // Repeat until trees API succeeds or dir_parts exhausted
        while !dir_parts.is_empty() {
            parsed_url.dir = dir_parts.join("/");

            match self.via_trees_api(parsed_url, token).await {
                Ok((tree_files, is_truncated)) => {
                    files = tree_files;
                    truncated = is_truncated;
                    break;
                }
                Err(RepoPackError::NotFound { .. }) => {
                    // Shift first dir part into ref (branch name extends)
                    parsed_url.git_ref = format!("{}/{}", parsed_url.git_ref, dir_parts[0]);
                    dir_parts = dir_parts[1..].to_vec();
                }
                Err(e) => return Err(e),
            }
        }

        // Trees API truncates at ~100k entries; fall back to slower contents API
        if files.is_empty() && truncated {
            files = self.via_contents_api(parsed_url, token).await?;
        }

        Ok(files)
    }

    /// Fetch file list via Git Trees API (fast, single request, but may truncate)
    async fn via_trees_api(
        &self,
        parsed_url: &ParsedUrl,
        token: Option<&str>,
    ) -> Result<(Vec<String>, bool), RepoPackError> {
        // Ensure dir ends with "/" for prefix matching (empty dir = repo root)
        let dir_prefix = if parsed_url.dir.is_empty() {
            String::new()
        } else if parsed_url.dir.ends_with('/') {
            parsed_url.dir.clone()
        } else {
            format!("{}/", parsed_url.dir)
        };

        let endpoint = format!(
            "{}/{}/git/trees/{}?recursive=1",
            parsed_url.owner, parsed_url.repo, parsed_url.git_ref
        );

        let response: TreeResponse = self.api_request(&endpoint, token).await?;

        // Filter to blobs (files) within the target directory
        let files: Vec<String> = response
            .tree
            .into_iter()
            .filter(|item| {
                item.item_type == "blob"
                    && (dir_prefix.is_empty() || item.path.starts_with(&dir_prefix))
            })
            .map(|item| item.path)
            .collect();

        Ok((files, response.truncated))
    }

    /// Fetch file list via Contents API (slower, recursive, but complete)
    async fn via_contents_api(
        &self,
        parsed_url: &ParsedUrl,
        token: Option<&str>,
    ) -> Result<Vec<String>, RepoPackError> {
        let endpoint = format!(
            "{}/{}/contents/{}?ref={}",
            parsed_url.owner, parsed_url.repo, parsed_url.dir, parsed_url.git_ref
        );

        let items: Vec<ContentItem> = self.api_request(&endpoint, token).await?;
        let mut files = Vec::new();

        for item in items {
            match item.item_type.as_str() {
                "file" => files.push(item.path),
                "dir" => {
                    // Recursively fetch subdirectory contents
                    let mut sub_url = parsed_url.clone();
                    sub_url.dir = item.path;
                    let sub_files = Box::pin(self.via_contents_api(&sub_url, token)).await?;
                    files.extend(sub_files);
                }
                _ => {} // Ignore symlinks, submodules, etc.
            }
        }

        Ok(files)
    }

    async fn api_request<T: serde::de::DeserializeOwned>(
        &self,
        endpoint: &str,
        token: Option<&str>,
    ) -> Result<T, RepoPackError> {
        let url = format!("https://api.github.com/repos/{endpoint}");

        let mut request = self.client.get(&url);

        if let Some(token) = token {
            request = request.header("Authorization", format!("Bearer {token}"));
        }

        let response = request
            .send()
            .await
            .map_err(|e| RepoPackError::DownloadFailed {
                path: endpoint.to_string(),
                source: e,
            })?;

        let status = response.status();

        // Rate limit: 60/hr unauthenticated, 5000/hr with token
        if status == StatusCode::FORBIDDEN {
            if let Some(remaining) = response.headers().get("X-RateLimit-Remaining")
                && remaining == "0"
            {
                let reset_time = response
                    .headers()
                    .get("X-RateLimit-Reset")
                    .and_then(|v| v.to_str().ok())
                    .unwrap_or("unknown")
                    .to_string();
                return Err(RepoPackError::RateLimited { reset_time });
            }
            return Err(RepoPackError::AuthRequired);
        }

        // Secondary rate limit (abuse detection)
        if status == StatusCode::TOO_MANY_REQUESTS {
            let reset_time = response
                .headers()
                .get("Retry-After")
                .and_then(|v| v.to_str().ok())
                .unwrap_or("unknown")
                .to_string();
            return Err(RepoPackError::RateLimited { reset_time });
        }

        if status == StatusCode::NOT_FOUND {
            let parts: Vec<&str> = endpoint.split('/').collect();
            return Err(RepoPackError::NotFound {
                owner: parts.first().unwrap_or(&"unknown").to_string(),
                repo: parts.get(1).unwrap_or(&"unknown").to_string(),
                hint: "Check that the repository exists and the URL is correct".to_string(),
            });
        }

        if let Err(err) = response.error_for_status_ref() {
            return Err(RepoPackError::DownloadFailed {
                path: endpoint.to_string(),
                source: err.without_url(),
            });
        }

        response
            .json::<T>()
            .await
            .map_err(|e| RepoPackError::DownloadFailed {
                path: endpoint.to_string(),
                source: e,
            })
    }

    /// Downloads a file from the repository, following Git LFS pointers if detected.
    ///
    /// Files are fetched from `raw.githubusercontent.com`. If the response matches the
    /// LFS pointer signature (128–140 bytes starting with the version header), the actual
    /// content is fetched from `media.githubusercontent.com`.
    pub async fn download_file(
        &self,
        path: &str,
        parsed_url: &ParsedUrl,
    ) -> Result<Bytes, RepoPackError> {
        let encoded_path = urlencoding::encode(path);
        let raw_url = format!(
            "https://raw.githubusercontent.com/{}/{}/{}/{}",
            parsed_url.owner, parsed_url.repo, parsed_url.git_ref, encoded_path
        );

        let response = self
            .client
            .get(&raw_url)
            .send()
            .await
            .map_err(|e| self.download_err(path, e))?;

        if let Err(err) = response.error_for_status_ref() {
            return Err(self.download_err(path, err.without_url()));
        }

        let content_length = response
            .headers()
            .get("content-length")
            .and_then(|v| v.to_str().ok())
            .and_then(|s| s.parse::<usize>().ok());

        // Fast path: content length outside LFS pointer range
        if !might_be_lfs_pointer(content_length) {
            return response
                .bytes()
                .await
                .map_err(|e| self.download_err(path, e));
        }

        // Content length suggests possible LFS pointer - need to check body
        let body = response
            .bytes()
            .await
            .map_err(|e| self.download_err(path, e))?;

        if !is_lfs_pointer(&body) {
            return Ok(body);
        }

        // LFS pointer detected - fetch actual content from media URL
        let lfs_url = format!(
            "https://media.githubusercontent.com/media/{}/{}/{}/{}",
            parsed_url.owner, parsed_url.repo, parsed_url.git_ref, encoded_path
        );

        let lfs_response = self
            .client
            .get(&lfs_url)
            .send()
            .await
            .map_err(|e| self.download_err(path, e))?;

        if let Err(err) = lfs_response.error_for_status_ref() {
            return Err(self.download_err(path, err.without_url()));
        }

        lfs_response
            .bytes()
            .await
            .map_err(|e| self.download_err(path, e))
    }
}

/// Check if content length suggests a possible LFS pointer (128-140 bytes).
#[inline]
fn might_be_lfs_pointer(content_length: Option<usize>) -> bool {
    content_length.is_some_and(|len| (128..=140).contains(&len))
}

/// Check if response body is a Git LFS pointer.
#[inline]
fn is_lfs_pointer(body: &[u8]) -> bool {
    const LFS_VERSION_PREFIX: &[u8] = b"version https://git-lfs.github.com/spec/v1";
    body.len() >= LFS_VERSION_PREFIX.len() && body.starts_with(LFS_VERSION_PREFIX)
}
