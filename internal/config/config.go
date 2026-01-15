package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Theme    string `json:"theme"`
	ASCIIArt string `json:"ascii_art"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Theme:    "default",
		ASCIIArt: "fire",
	}
}

// LoadConfig loads configuration from ~/.config/warpp/config.json
func LoadConfig() (Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return DefaultConfig(), err
	}

	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultConfig(), err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return DefaultConfig(), err
	}

	return config, nil
}

// SaveConfig saves configuration to ~/.config/warpp/config.json
func SaveConfig(config Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// getConfigPath returns the path to the config file
func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".config", "warpp", "config.json"), nil
}
