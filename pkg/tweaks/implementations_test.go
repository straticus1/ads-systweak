package tweaks

import (
	"context"
	"errors"
	"testing"

	"ads-systweak/pkg/execx"
)

type countingRunner struct {
	calls  int
	output string
	err    error
}

func (r *countingRunner) Run(context.Context, string, ...string) (string, error) {
	r.calls++
	return r.output, r.err
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

func TestDefaultsProbeTreatsMissingKeyAsOff(t *testing.T) {
	oldRunner := commandRunner
	t.Cleanup(func() { commandRunner = oldRunner })
	commandRunner = &countingRunner{err: &execx.CommandError{
		Name:   "/usr/bin/defaults",
		Stderr: "The domain/default pair of (com.example, Missing) does not exist",
		Err:    errors.New("exit status 1"),
	}}
	tweak := NewDefaultsTweak("test", "Test", "Test", CategoryOther, RiskLow, "com.example", "Missing", "bool", true, "")

	result := tweak.Probe()
	if result.State != ProbeOff || result.Err != nil {
		t.Fatalf("Probe = %#v, want off without error", result)
	}
}

func TestDefaultsProbeReportsPermissionDenied(t *testing.T) {
	oldRunner := commandRunner
	t.Cleanup(func() { commandRunner = oldRunner })
	commandRunner = &countingRunner{err: &execx.CommandError{
		Name:   "/usr/bin/defaults",
		Stderr: "Permission denied",
		Err:    errors.New("exit status 1"),
	}}
	tweak := NewDefaultsTweak("test", "Test", "Test", CategoryOther, RiskLow, "com.example", "Protected", "bool", true, "")

	result := tweak.Probe()
	if result.State != ProbePermissionDenied || result.Err == nil {
		t.Fatalf("Probe = %#v, want permission denied with error", result)
	}
}

func TestDefaultsProbeDoesNotTreatTrueAsAppliedForFalseTarget(t *testing.T) {
	oldRunner := commandRunner
	t.Cleanup(func() { commandRunner = oldRunner })
	commandRunner = &countingRunner{output: "1"}
	tweak := NewDefaultsTweak("test", "Test", "Test", CategoryOther, RiskLow, "com.example", "DisabledSetting", "bool", false, "")

	result := tweak.Probe()
	if result.State != ProbeOff || result.Applied {
		t.Fatalf("Probe = %#v, want off for actual true and target false", result)
	}
}

func TestCommandProbeReportsExecutionError(t *testing.T) {
	oldRunner := commandRunner
	t.Cleanup(func() { commandRunner = oldRunner })
	commandRunner = &countingRunner{err: errors.New("command missing")}
	tweak := NewCommandTweak("test", "Test", "Test", CategoryOther, RiskLow, "missing-command", "true", "true", false)

	result := tweak.Probe()
	if result.State != ProbeError || result.Err == nil {
		t.Fatalf("Probe = %#v, want error state", result)
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
