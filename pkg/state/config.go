package state

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	DesiredState map[string]bool   `json:"desired_state"`
	PreApproved  []string          `json:"pre_approved"`
}

func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ads-systweak.json")
}

func LoadConfig() (*Config, error) {
	path := GetConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				DesiredState: make(map[string]bool),
				PreApproved:  []string{},
			}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	
	if cfg.DesiredState == nil {
		cfg.DesiredState = make(map[string]bool)
	}
	if cfg.PreApproved == nil {
		cfg.PreApproved = []string{}
	}

	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	path := GetConfigPath()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func IsPreApproved(cfg *Config, tweakID string, category string) bool {
	for _, pa := range cfg.PreApproved {
		if pa == tweakID || pa == category {
			return true
		}
	}
	return false
}
