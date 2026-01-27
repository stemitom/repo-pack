use crate::error::RepoPackError;
use serde::{Deserialize, Serialize};
use std::path::PathBuf;

/// Application configuration for repo-pack.
///
/// Configuration is stored at `~/.config/repo-pack/config.json` and created with
/// default values on first run.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Config {
    pub concurrent_download_limit: u64,
    pub progress_bar_style: String,
    pub github_token_path: PathBuf,
}

impl Default for Config {
    fn default() -> Self {
        let home = dirs::home_dir().unwrap_or_else(|| PathBuf::from("~"));
        Self {
            concurrent_download_limit: 5,
            progress_bar_style: "█".to_string(),
            github_token_path: home.join(".github").join("token"),
        }
    }
}

impl Config {
    /// Loads configuration from disk, creating a default config file if none exists.
    pub fn load() -> Result<Self, RepoPackError> {
        let config_path = Self::config_path();

        if !config_path.exists() {
            let config = Config::default();
            config.save()?;
            return Ok(config);
        }

        let contents = std::fs::read_to_string(&config_path)
            .map_err(|e| RepoPackError::ConfigLoad { source: e })?;

        serde_json::from_str(&contents).map_err(|e| RepoPackError::ConfigParse { source: e })
    }

    /// Persists the current configuration to disk.
    pub fn save(&self) -> Result<(), RepoPackError> {
        let config_path = Self::config_path();

        if let Some(parent) = config_path.parent() {
            std::fs::create_dir_all(parent).map_err(|e| RepoPackError::ConfigSave { source: e })?;
        }

        let contents =
            serde_json::to_string_pretty(self).map_err(|e| RepoPackError::ConfigSave {
                source: std::io::Error::new(std::io::ErrorKind::InvalidData, e),
            })?;

        std::fs::write(&config_path, contents).map_err(|e| RepoPackError::ConfigSave { source: e })
    }

    fn config_path() -> PathBuf {
        dirs::config_dir()
            .unwrap_or_else(|| {
                dirs::home_dir()
                    .unwrap_or_else(|| PathBuf::from("."))
                    .join(".config")
            })
            .join("repo-pack")
            .join("config.json")
    }

    /// Reads the GitHub token from the configured `github_token_path`, if present.
    pub fn read_token(&self) -> Option<String> {
        std::fs::read_to_string(&self.github_token_path)
            .ok()
            .map(|s| s.trim().to_string())
            .filter(|s| !s.is_empty())
    }
}
