package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Config struct {
	DesiredState map[string]bool `json:"desired_state"`
	PreApproved  []string        `json:"pre_approved"`
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
	if cfg == nil {
		return errors.New("config cannot be nil")
	}
	path := GetConfigPath()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomic(path, append(data, '\n'), 0o600)
}

func writeAtomic(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	return os.Chmod(path, mode)
}

func IsPreApproved(cfg *Config, tweakID string, category string) bool {
	for _, pa := range cfg.PreApproved {
		if pa == tweakID || pa == category {
			return true
		}
	}
	return false
}
