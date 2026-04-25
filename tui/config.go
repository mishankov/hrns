package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ProviderConfig struct {
	Url        string `json:"url"`
	Key        string `json:"key"`
	Model      string `json:"model"`
	SkipVerify bool   `json:"skipVerify"`
}

type Config struct {
	Providers       map[string]ProviderConfig `json:"providers"`
	CurrentProvider string                    `json:"currentProvider"`
	CurrentAgent    string                    `json:"currentAgent"`
}

func LoadConfig() (*Config, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(userHome, ".config", "hrns")
	configPath := filepath.Join(configDir, "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) Save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	userHome, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(userHome, ".config", "hrns")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return err
	}

	return nil
}
