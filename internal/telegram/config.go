package telegram

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the Telegram bridge configuration.
type Config struct {
	Token  string `json:"token"`
	ChatID int64  `json:"chat_id"`
}

// LoadConfig loads the Telegram configuration from the town root.
func LoadConfig(townRoot string) (*Config, error) {
	path := filepath.Join(townRoot, "mayor", "telegram.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading telegram config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing telegram config: %w", err)
	}

	return &cfg, nil
}

// SaveConfig saves the Telegram configuration to the town root.
func SaveConfig(townRoot string, cfg *Config) error {
	path := filepath.Join(townRoot, "mayor", "telegram.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
