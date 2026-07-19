package state

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSaveConfigWritesPrivateFileAndRoundTrips(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	want := &Config{
		DesiredState: map[string]bool{"show-hidden-files": true},
		PreApproved:  []string{"System"},
	}
	if err := SaveConfig(want); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	info, err := os.Stat(GetConfigPath())
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o, want 600", info.Mode().Perm())
	}
	got, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("config = %#v, want %#v", got, want)
	}
}

func TestSaveConfigAtomicallyReplacesExistingFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := os.WriteFile(GetConfigPath(), []byte(`{"desired_state":{"old":true}}`), 0o644); err != nil {
		t.Fatalf("seed config: %v", err)
	}
	want := &Config{DesiredState: map[string]bool{"new": true}, PreApproved: []string{}}
	if err := SaveConfig(want); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	got, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("config = %#v, want %#v", got, want)
	}
	if matches, _ := filepathMatches(t, ".ads-systweak.json.tmp-*"); len(matches) != 0 {
		t.Fatalf("temporary files remain: %v", matches)
	}
}

func TestLoadConfigReportsMalformedJSON(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := os.WriteFile(GetConfigPath(), []byte(`{"desired_state":`), 0o600); err != nil {
		t.Fatalf("seed config: %v", err)
	}
	if _, err := LoadConfig(); err == nil {
		t.Fatal("LoadConfig returned nil error")
	}
}

func filepathMatches(t *testing.T, pattern string) ([]string, error) {
	t.Helper()
	return filepath.Glob(filepath.Join(filepath.Dir(GetConfigPath()), pattern))
}
