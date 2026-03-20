package backup

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// BackupEntry represents a single defaults state that was backed up
type BackupEntry struct {
	Domain    string `json:"domain"`
	Key       string `json:"key"`
	Value     string `json:"value"` // Store raw output of `defaults read`
	Exists    bool   `json:"exists"` // True if it existed before getting overwritten
	Timestamp string `json:"timestamp"`
}

func runShell(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return strings.TrimSpace(out.String()), err
}

// EnsureBackupDir creates the backup directory in the user's home folder
func EnsureBackupDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".ads-systweak-backups")
	err = os.MkdirAll(dir, 0755)
	return dir, err
}

// SaveBackup reads the current default value (if any) and appends it to a backup log
func SaveBackup(domain, key string) error {
	dir, err := EnsureBackupDir()
	if err != nil {
		return err
	}

	val, err := runShell(fmt.Sprintf("defaults read %s %s", domain, key))
	exists := err == nil

	entry := BackupEntry{
		Domain:    domain,
		Key:       key,
		Value:     val,
		Exists:    exists,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	b, _ := json.Marshal(entry)
	
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(domain+key)))
	file := filepath.Join(dir, hash+".json")
	return os.WriteFile(file, b, 0644)
}

// RestoreAll reads all backup files in the backup directory and restores them
func RestoreAll() error {
	dir, err := EnsureBackupDir()
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}

		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}

		var entry BackupEntry
		if err := json.Unmarshal(b, &entry); err != nil {
			continue
		}

		if !entry.Exists {
			runShell(fmt.Sprintf("defaults delete %s %s", entry.Domain, entry.Key))
		} else {
			runShell(fmt.Sprintf("defaults write %s %s -string \"%s\"", entry.Domain, entry.Key, strings.ReplaceAll(entry.Value, "\"", "\\\"")))
		}
	}
	return nil
}
