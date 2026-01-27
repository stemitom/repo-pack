use repo_pack::Config;
use std::fs;
use std::io::Write;

#[test]
fn config_default_values() {
    let config = Config::default();
    assert_eq!(config.concurrent_download_limit, 5);
    assert_eq!(config.progress_bar_style, "█");
}

#[test]
fn config_read_token_returns_none_if_missing() {
    let temp_dir = tempfile::tempdir().unwrap();
    let config = Config {
        concurrent_download_limit: 5,
        progress_bar_style: "█".to_string(),
        github_token_path: temp_dir.path().join("nonexistent_token"),
    };

    assert!(config.read_token().is_none());
}

#[test]
fn config_read_token_returns_trimmed_content() {
    let temp_dir = tempfile::tempdir().unwrap();
    let token_path = temp_dir.path().join("token");

    let mut file = fs::File::create(&token_path).unwrap();
    writeln!(file, "  ghp_test_token_123  ").unwrap();

    let config = Config {
        concurrent_download_limit: 5,
        progress_bar_style: "█".to_string(),
        github_token_path: token_path,
    };

    assert_eq!(config.read_token(), Some("ghp_test_token_123".to_string()));
}

#[test]
fn config_read_token_returns_none_for_empty_file() {
    let temp_dir = tempfile::tempdir().unwrap();
    let token_path = temp_dir.path().join("token");

    fs::File::create(&token_path).unwrap();

    let config = Config {
        concurrent_download_limit: 5,
        progress_bar_style: "█".to_string(),
        github_token_path: token_path,
    };

    assert!(config.read_token().is_none());
}

#[test]
fn config_read_token_returns_none_for_whitespace_only() {
    let temp_dir = tempfile::tempdir().unwrap();
    let token_path = temp_dir.path().join("token");

    let mut file = fs::File::create(&token_path).unwrap();
    writeln!(file, "   \n\t  ").unwrap();

    let config = Config {
        concurrent_download_limit: 5,
        progress_bar_style: "█".to_string(),
        github_token_path: token_path,
    };

    assert!(config.read_token().is_none());
}
