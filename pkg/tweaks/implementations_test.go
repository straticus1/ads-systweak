package tweaks

import (
	"context"
	"errors"
	"testing"
)

type countingRunner struct{ calls int }

func (r *countingRunner) Run(context.Context, string, ...string) (string, error) {
	r.calls++
	return "", nil
}

func TestDefaultsApplyStopsWhenBackupFails(t *testing.T) {
	oldSave := saveDefaultsBackup
	oldRunner := commandRunner
	t.Cleanup(func() {
		saveDefaultsBackup = oldSave
		commandRunner = oldRunner
	})

	saveDefaultsBackup = func(string, string) error { return errors.New("backup failed") }
	runner := &countingRunner{}
	commandRunner = runner
	tweak := NewDefaultsTweak("test", "Test", "Test", CategoryOther, RiskLow, "com.example", "Enabled", "bool", true, "")

	if err := tweak.Apply(); err == nil {
		t.Fatal("Apply returned nil error")
	}
	if runner.calls != 0 {
		t.Fatalf("runner calls = %d, want 0", runner.calls)
	}
}

func TestDefaultsRevertUsesBackupRestore(t *testing.T) {
	oldRestore := restoreDefaultsBackup
	oldRunner := commandRunner
	t.Cleanup(func() {
		restoreDefaultsBackup = oldRestore
		commandRunner = oldRunner
	})

	restored := false
	restoreDefaultsBackup = func(domain, key string) error {
		restored = domain == "com.example" && key == "Enabled"
		return nil
	}
	runner := &countingRunner{}
	commandRunner = runner
	tweak := NewDefaultsTweak("test", "Test", "Test", CategoryOther, RiskLow, "com.example", "Enabled", "bool", true, "")

	if err := tweak.Revert(); err != nil {
		t.Fatalf("Revert: %v", err)
	}
	if !restored {
		t.Fatal("backup restore was not called")
	}
	if runner.calls != 0 {
		t.Fatalf("runner calls = %d, want 0", runner.calls)
	}
}
