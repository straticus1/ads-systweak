package cmd

import (
	"strings"
	"testing"
)

func TestPlatformWarningOnlyWarnsOutsideMacOS(t *testing.T) {
	if got := platformWarning("darwin"); got != "" {
		t.Fatalf("darwin warning = %q", got)
	}
	if got := platformWarning("linux"); !strings.Contains(got, "linux") || !strings.Contains(got, "macOS") {
		t.Fatalf("linux warning = %q", got)
	}
}
