use crate::error::RepoPackError;
use std::borrow::Cow;
use std::path::{Component, Path, PathBuf};

/// Check if a path needs normalization (contains `.` or `..` components).
fn needs_normalization(path: &Path) -> bool {
    path.components()
        .any(|c| matches!(c, Component::ParentDir | Component::CurDir))
}

/// Normalize a path by resolving `.` and `..` components.
/// Returns `Cow::Borrowed` if no normalization needed, `Cow::Owned` otherwise.
fn normalize_path(path: &Path) -> Cow<'_, Path> {
    if !needs_normalization(path) {
        return Cow::Borrowed(path);
    }

    let mut normalized = PathBuf::new();
    for component in path.components() {
        match component {
            Component::Prefix(_) | Component::RootDir | Component::Normal(_) => {
                normalized.push(component);
            }
            Component::ParentDir => {
                // Pop the last component if it's a normal component
                if matches!(
                    normalized.components().next_back(),
                    Some(Component::Normal(_))
                ) {
                    normalized.pop();
                } else {
                    // Preserve leading `..` or `..` after root
                    normalized.push(component);
                }
            }
            Component::CurDir => {
                // Skip `.` components
            }
        }
    }
    Cow::Owned(normalized)
}

/// Extracts the relative path starting from `base_dir`.
///
/// Given a full `file_path`, this function locates the `base_dir` component and returns
/// the path from that point onwards — preserving the directory structure.
///
/// For example, `"path/to/nvim/lua/config.lua"` with base `"nvim"` yields `"nvim/lua/config.lua"`.
pub fn extract_relative_path(base_dir: &str, file_path: &str) -> Result<String, RepoPackError> {
    let base_path = Path::new(base_dir);
    let file_path_obj = Path::new(file_path);

    let (base_dir, file_path) =
        if needs_normalization(base_path) || needs_normalization(file_path_obj) {
            let normalized_base = base_path
                .components()
                .collect::<PathBuf>()
                .to_string_lossy()
                .to_string();
            let normalized_file = file_path_obj
                .components()
                .collect::<PathBuf>()
                .to_string_lossy()
                .to_string();
            (normalized_base, normalized_file)
        } else {
            (base_dir.to_string(), file_path.to_string())
        };

    // GitHub API always returns paths with forward slashes
    let search_pattern = format!("{base_dir}/");

    if let Some(index) = file_path.find(&search_pattern) {
        return Ok(file_path[index..].to_string());
    }

    if file_path.ends_with(&base_dir) {
        return Ok(String::new());
    }

    Err(RepoPackError::PathTraversal {
        path: format!("base directory {base_dir} not found in file path {file_path}"),
    })
}

/// Saves file content to `output_dir` with path traversal protection.
///
/// The relative path is extracted from `file_path` using `base_dir` as the anchor point.
/// Before writing, the resolved path is validated to ensure it remains within `output_dir`
/// bounds — rejecting any `..` sequences that would escape the output directory.
pub async fn save_file(
    base_dir: &str,
    file_path: &str,
    content: &[u8],
    output_dir: &Path,
) -> Result<PathBuf, RepoPackError> {
    let adjusted_file_path = extract_relative_path(base_dir, file_path)?;

    let full_path = output_dir.join(&adjusted_file_path);
    let full_path = normalize_path(&full_path);
    let normalized_output_dir = normalize_path(output_dir);

    let output_components: Vec<_> = normalized_output_dir.components().collect();
    let full_components: Vec<_> = full_path.components().collect();

    // Check if output_dir components are a prefix of full_path components
    if full_components.len() < output_components.len()
        || !full_components
            .iter()
            .zip(output_components.iter())
            .all(|(a, b)| a == b)
    {
        return Err(RepoPackError::PathTraversal {
            path: format!(
                "{} is outside output directory {}",
                file_path,
                output_dir.display()
            ),
        });
    }

    let full_path = full_path.into_owned();

    if let Some(parent) = full_path.parent() {
        fs_err::tokio::create_dir_all(parent)
            .await
            .map_err(|source| RepoPackError::IoError {
                path: parent.to_path_buf(),
                source,
            })?;
    }

    fs_err::tokio::write(&full_path, content)
        .await
        .map_err(|source| RepoPackError::IoError {
            path: full_path.clone(),
            source,
        })?;

    Ok(full_path)
}

/// Result of a batch download operation.
#[derive(Debug, Default)]
pub struct DownloadResult {
    pub downloaded: u64,
    pub skipped: u64,
    pub failed: u64,
    pub cancelled: bool,
    pub errors: Vec<(String, RepoPackError)>,
}

/// Options for the download operation.
pub struct DownloadOptions<'a> {
    pub base_dir: &'a str,
    pub output_dir: &'a Path,
    pub concurrency_limit: usize,
    pub resume: bool,
    pub verbose: bool,
}

use crate::progress::DownloadProgress;
use crate::provider::GitHubProvider;
use crate::url::ParsedUrl;
use futures::stream::{self, StreamExt};
use std::sync::Arc;
use std::sync::atomic::{AtomicBool, Ordering};
use tokio::signal;
use tokio::sync::Semaphore;

/// Cancellation token for cooperative shutdown.
pub type CancellationToken = Arc<AtomicBool>;

/// Downloads multiple files concurrently with progress reporting.
///
/// Uses a semaphore to limit concurrent downloads. If `resume` is enabled,
/// existing files are skipped. Checks `cancelled` token between downloads
/// for cooperative shutdown. Returns aggregate results including any errors.
pub async fn download_files(
    provider: &GitHubProvider,
    parsed_url: &ParsedUrl,
    files: Vec<String>,
    options: DownloadOptions<'_>,
    progress: &DownloadProgress,
    cancelled: &CancellationToken,
) -> DownloadResult {
    let semaphore = Arc::new(Semaphore::new(options.concurrency_limit));
    let cancelled_inner = cancelled.clone();
    let mut result = DownloadResult::default();

    let tasks: Vec<_> = files
        .into_iter()
        .map(|file_path| {
            let semaphore = semaphore.clone();
            let cancelled = cancelled_inner.clone();
            let base_dir = options.base_dir.to_string();
            let output_dir = options.output_dir.to_path_buf();
            let resume = options.resume;

            async move {
                if cancelled.load(Ordering::Relaxed) {
                    return (file_path, DownloadStatus::Cancelled);
                }

                let _permit = semaphore.acquire().await.expect("semaphore closed");

                if cancelled.load(Ordering::Relaxed) {
                    return (file_path, DownloadStatus::Cancelled);
                }

                if resume
                    && let Ok(relative) = extract_relative_path(&base_dir, &file_path)
                    && output_dir.join(&relative).exists()
                {
                    return (file_path, DownloadStatus::Skipped);
                }

                match provider.download_file(&file_path, parsed_url).await {
                    Ok(content) => {
                        match save_file(&base_dir, &file_path, &content, &output_dir).await {
                            Ok(_) => (file_path, DownloadStatus::Downloaded),
                            Err(e) => (file_path, DownloadStatus::Failed(e)),
                        }
                    }
                    Err(e) => (file_path, DownloadStatus::Failed(e)),
                }
            }
        })
        .collect();

    let mut task_stream = stream::iter(tasks).buffer_unordered(options.concurrency_limit);
    let mut ctrl_c = std::pin::pin!(signal::ctrl_c());

    loop {
        tokio::select! {
            biased;

            _ = &mut ctrl_c => {
                cancelled.store(true, Ordering::SeqCst);
                result.cancelled = true;
                progress.close();
                break;
            }

            task_result = task_stream.next() => {
                match task_result {
                    Some((file_path, status)) => {
                        match status {
                            DownloadStatus::Downloaded => {
                                result.downloaded += 1;
                                progress.set_current_file(&file_path);
                                progress.inc();
                            }
                            DownloadStatus::Skipped => {
                                result.skipped += 1;
                                progress.inc();
                            }
                            DownloadStatus::Failed(e) => {
                                result.failed += 1;
                                result.errors.push((file_path, e));
                                progress.inc();
                            }
                            DownloadStatus::Cancelled => {}
                        }
                    }
                    None => {
                        progress.close();
                        break;
                    }
                }
            }
        }
    }

    result
}

enum DownloadStatus {
    Downloaded,
    Skipped,
    Failed(RepoPackError),
    Cancelled,
}
