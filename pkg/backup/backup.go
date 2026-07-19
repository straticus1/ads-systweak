package backup

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ads-systweak/pkg/execx"
)

var (
	ErrNotFound        = errors.New("defaults key not found")
	ErrNoBackup        = errors.New("no backup receipt")
	ErrUnsupportedType = errors.New("unsupported defaults value type")
)

// ValueType is a defaults scalar type that can be restored without coercion.
type ValueType string

const (
	TypeBool   ValueType = "bool"
	TypeInt    ValueType = "int"
	TypeFloat  ValueType = "float"
	TypeString ValueType = "string"
)

// BackupEntry is an exact, typed receipt captured before a preference mutation.
type BackupEntry struct {
	Domain    string    `json:"domain"`
	Key       string    `json:"key"`
	Value     string    `json:"value,omitempty"`
	Type      ValueType `json:"type,omitempty"`
	Exists    bool      `json:"exists"`
	Timestamp string    `json:"timestamp"`
}

// DefaultsClient reads and restores typed values.
type DefaultsClient interface {
	Read(domain, key string) (string, ValueType, error)
	Write(domain, key string, valueType ValueType, value string) error
	Delete(domain, key string) error
}

// Store manages rollback receipts in a directory.
type Store struct {
	Dir    string
	Client DefaultsClient
	Now    func() time.Time
}

// SaveIfAbsent captures the first pre-mutation value and never overwrites it.
func (s *Store) SaveIfAbsent(domain, key string) error {
	if err := s.validate(); err != nil {
		return err
	}
	if _, err := os.Stat(s.path(domain, key)); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect backup receipt: %w", err)
	}

	value, valueType, err := s.Client.Read(domain, key)
	entry := BackupEntry{Domain: domain, Key: key, Timestamp: s.Now().Format(time.RFC3339Nano)}
	if err == nil {
		if !supported(valueType) {
			return fmt.Errorf("%w: %s", ErrUnsupportedType, valueType)
		}
		entry.Exists = true
		entry.Value = value
		entry.Type = valueType
	} else if !errors.Is(err, ErrNotFound) {
		return fmt.Errorf("read original defaults value: %w", err)
	}

	return writeJSONAtomic(s.path(domain, key), entry)
}

// Load reads a rollback receipt.
func (s *Store) Load(domain, key string) (BackupEntry, error) {
	var entry BackupEntry
	data, err := os.ReadFile(s.path(domain, key))
	if os.IsNotExist(err) {
		return entry, ErrNoBackup
	}
	if err != nil {
		return entry, fmt.Errorf("read backup receipt: %w", err)
	}
	if err := json.Unmarshal(data, &entry); err != nil {
		return entry, fmt.Errorf("decode backup receipt: %w", err)
	}
	if entry.Domain != domain || entry.Key != key {
		return entry, errors.New("backup receipt identity mismatch")
	}
	return entry, nil
}

// Restore restores one value and consumes its receipt only after success.
func (s *Store) Restore(domain, key string) error {
	if err := s.validate(); err != nil {
		return err
	}
	entry, err := s.Load(domain, key)
	if err != nil {
		return err
	}
	if entry.Exists {
		if !supported(entry.Type) {
			return fmt.Errorf("%w: %s", ErrUnsupportedType, entry.Type)
		}
		err = s.Client.Write(domain, key, entry.Type, entry.Value)
	} else {
		err = s.Client.Delete(domain, key)
	}
	if err != nil {
		return fmt.Errorf("restore defaults value: %w", err)
	}
	if err := os.Remove(s.path(domain, key)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("consume backup receipt: %w", err)
	}
	return nil
}

func (s *Store) validate() error {
	if s.Dir == "" || s.Client == nil || s.Now == nil {
		return errors.New("backup store is not configured")
	}
	return os.MkdirAll(s.Dir, 0o700)
}

func (s *Store) path(domain, key string) string {
	hash := sha256.Sum256([]byte(domain + "\x00" + key))
	return filepath.Join(s.Dir, fmt.Sprintf("%x.json", hash))
}

func supported(valueType ValueType) bool {
	switch valueType {
	case TypeBool, TypeInt, TypeFloat, TypeString:
		return true
	default:
		return false
	}
}

func writeJSONAtomic(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".backup-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if err := json.NewEncoder(tmp).Encode(value); err != nil {
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
	return os.Rename(tmpName, path)
}

type commandDefaultsClient struct{ runner execx.Runner }

func (c commandDefaultsClient) Read(domain, key string) (string, ValueType, error) {
	typeOutput, err := c.runner.Run(context.Background(), "/usr/bin/defaults", "read-type", domain, key)
	if err != nil {
		if isMissing(err) {
			return "", "", ErrNotFound
		}
		return "", "", err
	}
	valueType, err := parseType(typeOutput)
	if err != nil {
		return "", "", err
	}
	value, err := c.runner.Run(context.Background(), "/usr/bin/defaults", "read", domain, key)
	return value, valueType, err
}

func (c commandDefaultsClient) Write(domain, key string, valueType ValueType, value string) error {
	_, err := c.runner.Run(context.Background(), "/usr/bin/defaults", "write", domain, key, "-"+string(valueType), value)
	return err
}

func (c commandDefaultsClient) Delete(domain, key string) error {
	_, err := c.runner.Run(context.Background(), "/usr/bin/defaults", "delete", domain, key)
	if isMissing(err) {
		return nil
	}
	return err
}

func parseType(output string) (ValueType, error) {
	lower := strings.ToLower(output)
	switch {
	case strings.Contains(lower, "boolean"):
		return TypeBool, nil
	case strings.Contains(lower, "integer"):
		return TypeInt, nil
	case strings.Contains(lower, "float") || strings.Contains(lower, "double"):
		return TypeFloat, nil
	case strings.Contains(lower, "string"):
		return TypeString, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedType, strings.TrimSpace(output))
	}
}

func isMissing(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "does not exist") || strings.Contains(text, "not found")
}

func defaultStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return &Store{
		Dir:    filepath.Join(home, ".ads-systweak-backups"),
		Client: commandDefaultsClient{runner: execx.OSRunner{}},
		Now:    time.Now,
	}, nil
}

// SaveBackup preserves compatibility with existing callers while using first-write semantics.
func SaveBackup(domain, key string) error {
	store, err := defaultStore()
	if err != nil {
		return err
	}
	return store.SaveIfAbsent(domain, key)
}

// Restore restores a single preference receipt.
func Restore(domain, key string) error {
	store, err := defaultStore()
	if err != nil {
		return err
	}
	return store.Restore(domain, key)
}

// RestoreAll restores every valid receipt and reports all failures.
func RestoreAll() error {
	store, err := defaultStore()
	if err != nil {
		return err
	}
	if err := store.validate(); err != nil {
		return err
	}
	entries, err := os.ReadDir(store.Dir)
	if err != nil {
		return err
	}
	var restoreErrors []error
	for _, item := range entries {
		if item.IsDir() || filepath.Ext(item.Name()) != ".json" {
			continue
		}
		data, readErr := os.ReadFile(filepath.Join(store.Dir, item.Name()))
		if readErr != nil {
			restoreErrors = append(restoreErrors, readErr)
			continue
		}
		var entry BackupEntry
		if decodeErr := json.Unmarshal(data, &entry); decodeErr != nil {
			restoreErrors = append(restoreErrors, decodeErr)
			continue
		}
		if restoreErr := store.Restore(entry.Domain, entry.Key); restoreErr != nil {
			restoreErrors = append(restoreErrors, restoreErr)
		}
	}
	return errors.Join(restoreErrors...)
}
