package config

import (
	"fmt"
)

// InitConfig creates a default config file
func InitConfig() error {
	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("could not get config path: %w", err)
	}

	defaultConfig := DefaultConfig()
	if err := SaveConfig(defaultConfig); err != nil {
		return fmt.Errorf("could not save config: %w", err)
	}

	fmt.Printf("Created config file at: %s\n", configPath)
	return nil
}

// ShowConfigPath prints the path to the config file
func ShowConfigPath() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	fmt.Printf("Config file location: %s\n", configPath)
	return nil
}

// ShowConfig prints current configuration
func ShowConfig() error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("could not load config: %w", err)
	}

	configPath, _ := getConfigPath()
	fmt.Printf("Config file: %s\n", configPath)
	fmt.Printf("Theme: %s\n", config.Theme)
	fmt.Printf("ASCII Art: %s\n", config.ASCIIArt)
	return nil
}
