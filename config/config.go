package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	ConcurrentDownloadLimit int    `json:"concurrent_download_limit"`
	ProgressBarStyle        string `json:"progress_bar_style"`
	GithubTokenPath         string `json:"github_token_path"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "~"
	}
	return Config{
		ConcurrentDownloadLimit: 5,
		ProgressBarStyle:        "█",
		GithubTokenPath:         filepath.Join(homeDir, ".github", "token"),
	}
}

// LoadConfig loads the configuration from the config file
func LoadConfig() (Config, error) {
	configPath := getConfigPath()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return createDefaultConfig()
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("error reading config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("error parsing config file: %v", err)
	}

	return config, nil
}

// SaveConfig saves the configuration to the config file
func SaveConfig(config Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	configPath := getConfigPath()
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("error creating config directory: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}

	return nil
}

// getConfigPath returns the path to the config file
func getConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(configDir, "repo-pack", "config.json")
}

// createDefaultConfig creates a new config file with default values
func createDefaultConfig() (Config, error) {
	config := DefaultConfig()
	if err := SaveConfig(config); err != nil {
		return Config{}, fmt.Errorf("error creating default config: %v", err)
	}
	return config, nil
}
