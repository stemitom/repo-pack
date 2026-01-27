use repo_pack::download::{extract_relative_path, save_file};
use repo_pack::error::RepoPackError;

#[test]
fn extract_relative_path_with_nested_directory() {
    let result = extract_relative_path("config", "repo/src/config/settings/app.toml");
    assert_eq!(result.unwrap(), "config/settings/app.toml");
}

#[test]
fn extract_relative_path_single_file() {
    let result = extract_relative_path("src", "project/src/main.rs");
    assert_eq!(result.unwrap(), "src/main.rs");
}

#[test]
fn extract_relative_path_deep_nesting() {
    let result = extract_relative_path("lua", "dotfiles/.config/nvim/lua/plugins/init.lua");
    assert_eq!(result.unwrap(), "lua/plugins/init.lua");
}

#[test]
fn extract_relative_path_base_not_found() {
    let result = extract_relative_path("missing", "path/to/some/file.txt");
    assert!(matches!(result, Err(RepoPackError::PathTraversal { .. })));
}

#[test]
fn extract_relative_path_empty_after_base() {
    let result = extract_relative_path("nvim", "dotfiles/nvim");
    assert_eq!(result.unwrap(), "");
}

#[tokio::test]
async fn save_file_creates_nested_directories() {
    let temp_dir = tempfile::tempdir().unwrap();
    let output_dir = temp_dir.path();

    let result = save_file(
        "config",
        "repo/config/nested/deep/settings.json",
        b"{\"key\": \"value\"}",
        output_dir,
    )
    .await;

    assert!(result.is_ok());
    let saved_path = result.unwrap();
    assert!(saved_path.exists());
    assert_eq!(
        saved_path,
        output_dir.join("config/nested/deep/settings.json")
    );
}

#[tokio::test]
async fn save_file_rejects_path_traversal_parent_dir() {
    let temp_dir = tempfile::tempdir().unwrap();
    let output_dir = temp_dir.path();

    let result = save_file("base", "base/../../../etc/passwd", b"malicious", output_dir).await;

    assert!(matches!(result, Err(RepoPackError::PathTraversal { .. })));
}

#[tokio::test]
async fn save_file_rejects_path_traversal_hidden_in_middle() {
    let temp_dir = tempfile::tempdir().unwrap();
    let output_dir = temp_dir.path();

    let result = save_file(
        "src",
        "src/valid/../../../escape/file.txt",
        b"malicious",
        output_dir,
    )
    .await;

    assert!(matches!(result, Err(RepoPackError::PathTraversal { .. })));
}

#[tokio::test]
async fn save_file_allows_dot_in_filename() {
    let temp_dir = tempfile::tempdir().unwrap();
    let output_dir = temp_dir.path();

    let result = save_file("config", "config/.env.local", b"SECRET=value", output_dir).await;

    assert!(result.is_ok());
    let saved_path = result.unwrap();
    assert!(saved_path.exists());
    assert_eq!(saved_path, output_dir.join("config/.env.local"));
}

#[tokio::test]
async fn save_file_preserves_file_content() {
    let temp_dir = tempfile::tempdir().unwrap();
    let output_dir = temp_dir.path();
    let content = b"line 1\nline 2\nline 3\n";

    let result = save_file("data", "data/test.txt", content, output_dir).await;

    assert!(result.is_ok());
    let saved_path = result.unwrap();
    let read_content = fs_err::read(&saved_path).unwrap();
    assert_eq!(read_content, content);
}

#[tokio::test]
async fn save_file_handles_binary_content() {
    let temp_dir = tempfile::tempdir().unwrap();
    let output_dir = temp_dir.path();
    let binary_content: Vec<u8> = (0..=255).collect();

    let result = save_file("bin", "bin/data.bin", &binary_content, output_dir).await;

    assert!(result.is_ok());
    let saved_path = result.unwrap();
    let read_content = fs_err::read(&saved_path).unwrap();
    assert_eq!(read_content, binary_content);
}
