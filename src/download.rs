use crate::error::RepoPackError;
use std::path::{Path, PathBuf};

/// Extracts the relative path starting from base_dir.
///
/// This function finds the base_dir component in the file_path and returns
/// the path starting from that component onwards.
///
/// # Arguments
/// * `base_dir` - The base directory name to search for (last segment only)
/// * `file_path` - The full file path to extract from
///
/// # Returns
/// * `Ok(String)` - The relative path starting from base_dir
/// * `Err(RepoPackError)` - If base_dir is not found in file_path
///
/// # Example
/// ```
/// let result = extract_relative_path("nvim", "path/to/nvim/lua/config.lua");
/// assert_eq!(result.unwrap(), "nvim/lua/config.lua");
/// ```
pub fn extract_relative_path(base_dir: &str, file_path: &str) -> Result<String, RepoPackError> {
    // Normalize both paths by converting to Path and back to string
    let base_dir = Path::new(base_dir)
        .components()
        .collect::<PathBuf>()
        .to_string_lossy()
        .to_string();
    let file_path = Path::new(file_path)
        .components()
        .collect::<PathBuf>()
        .to_string_lossy()
        .to_string();

    // Look for baseDir as a path component (with separator after it)
    let separator = std::path::MAIN_SEPARATOR.to_string();
    let search_pattern = format!("{base_dir}{separator}");

    if let Some(index) = file_path.find(&search_pattern) {
        return Ok(file_path[index..].to_string());
    }

    // Try without separator at the end for exact match at end
    if file_path.ends_with(&base_dir) {
        return Ok(String::new());
    }

    Err(RepoPackError::PathTraversal {
        path: format!("base directory {base_dir} not found in file path {file_path}"),
    })
}

/// Saves a file to the output directory with path traversal protection.
///
/// This function:
/// 1. Extracts the relative path from base_dir
/// 2. Joins it with output_dir
/// 3. Verifies the result stays within output_dir bounds (path traversal check)
/// 4. Creates parent directories if needed
/// 5. Writes the file content
///
/// # Arguments
/// * `base_dir` - The base directory name to extract relative path from
/// * `file_path` - The full file path
/// * `content` - The file content as bytes
/// * `output_dir` - The output directory where files should be saved
///
/// # Returns
/// * `Ok(PathBuf)` - The full path where the file was saved
/// * `Err(RepoPackError)` - If path traversal is detected or IO error occurs
pub async fn save_file(
    base_dir: &str,
    file_path: &str,
    content: &[u8],
    output_dir: &Path,
) -> Result<PathBuf, RepoPackError> {
    // Extract relative path
    let adjusted_file_path = extract_relative_path(base_dir, file_path)?;

    // Join with output directory and normalize
    let full_path = output_dir.join(&adjusted_file_path);
    let full_path = full_path
        .components()
        .collect::<PathBuf>();

    // Path traversal protection: verify full_path is within output_dir
    // We need to compare normalized paths component by component
    let output_components: Vec<_> = output_dir.components().collect();
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

    // Create parent directories
    if let Some(parent) = full_path.parent() {
        fs_err::tokio::create_dir_all(parent)
            .await
            .map_err(|source| RepoPackError::IoError {
                path: parent.to_path_buf(),
                source,
            })?;
    }

    // Write file content
    fs_err::tokio::write(&full_path, content)
        .await
        .map_err(|source| RepoPackError::IoError {
            path: full_path.clone(),
            source,
        })?;

    Ok(full_path)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_extract_relative_path_basic() {
        let result = extract_relative_path("nvim", "path/to/nvim/lua/config.lua");
        assert_eq!(result.unwrap(), "nvim/lua/config.lua");
    }

    #[test]
    fn test_extract_relative_path_at_end() {
        let result = extract_relative_path("nvim", "path/to/nvim");
        assert_eq!(result.unwrap(), "");
    }

    #[test]
    fn test_extract_relative_path_not_found() {
        let result = extract_relative_path("nvim", "path/to/other/config.lua");
        assert!(result.is_err());
    }

    #[tokio::test]
    async fn test_save_file_basic() {
        let temp_dir = tempfile::tempdir().unwrap();
        let output_dir = temp_dir.path();

        let result = save_file(
            "nvim",
            "path/to/nvim/lua/config.lua",
            b"test content",
            output_dir,
        )
        .await;

        assert!(result.is_ok());
        let saved_path = result.unwrap();
        assert!(saved_path.exists());
        assert_eq!(
            fs_err::tokio::read_to_string(&saved_path).await.unwrap(),
            "test content"
        );
    }

    #[tokio::test]
    async fn test_save_file_path_traversal() {
        let temp_dir = tempfile::tempdir().unwrap();
        let output_dir = temp_dir.path();

        // Try to escape using ../
        let result = save_file(
            "nvim",
            "nvim/../../etc/passwd",
            b"malicious",
            output_dir,
        )
        .await;

        assert!(matches!(result, Err(RepoPackError::PathTraversal { .. })));
    }
}
