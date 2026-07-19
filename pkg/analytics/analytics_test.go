package analytics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetOrCreateUserIDIsPrivateStableAndNotHostnameDerived(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	hostname, _ := os.Hostname()

	first := getOrCreateUserID()
	second := getOrCreateUserID()
	if first == "" || first != second {
		t.Fatalf("IDs = %q and %q, want stable non-empty value", first, second)
	}
	if strings.Contains(first, hostname) {
		t.Fatalf("ID %q exposes hostname %q", first, hostname)
	}
	info, err := os.Stat(filepath.Join(home, ".ads-systweak-uid"))
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o, want 600", info.Mode().Perm())
	}
}

func TestGenerateUserIDProducesIndependentRandomValues(t *testing.T) {
	first := generateUserID()
	second := generateUserID()
	if first == "" || first == second {
		t.Fatalf("generated IDs = %q and %q", first, second)
	}
}
