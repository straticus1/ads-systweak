package tweaks

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type queueRunner struct {
	outputs []string
	errs    []error
	calls   [][]string
}

func (r *queueRunner) Run(_ context.Context, name string, args ...string) (string, error) {
	r.calls = append(r.calls, append([]string{name}, args...))
	index := len(r.calls) - 1
	var output string
	var err error
	if index < len(r.outputs) {
		output = r.outputs[index]
	}
	if index < len(r.errs) {
		err = r.errs[index]
	}
	return output, err
}

func TestAddBootArgumentPreservesExistingArguments(t *testing.T) {
	got := AddBootArgument("keepsyms=1 debug=0x100", "-v")
	if got != "keepsyms=1 debug=0x100 -v" {
		t.Fatalf("AddBootArgument = %q", got)
	}
	if duplicate := AddBootArgument(got, "-v"); duplicate != got {
		t.Fatalf("duplicate AddBootArgument = %q, want %q", duplicate, got)
	}
}

func TestRemoveBootArgumentPreservesOtherArguments(t *testing.T) {
	got := RemoveBootArgument("keepsyms=1 -v debug=0x100", "-v")
	if got != "keepsyms=1 debug=0x100" {
		t.Fatalf("RemoveBootArgument = %q", got)
	}
}

func TestParseNVRAMBootArgs(t *testing.T) {
	got := ParseNVRAMBootArgs("boot-args\tkeepsyms=1 -v")
	if got != "keepsyms=1 -v" {
		t.Fatalf("ParseNVRAMBootArgs = %q", got)
	}
}

func TestBootArgumentBuildersPreserveAndRestoreExactState(t *testing.T) {
	apply, err := BootArgumentApplyBuilder("-v")("boot-args\tkeepsyms=1")
	if err != nil {
		t.Fatalf("apply builder: %v", err)
	}
	if apply != "nvram boot-args='keepsyms=1 -v'" {
		t.Fatalf("apply = %q", apply)
	}
	restore, err := BootArgumentRestoreBuilder("boot-args\tkeepsyms=1")
	if err != nil {
		t.Fatalf("restore builder: %v", err)
	}
	if restore != "nvram boot-args='keepsyms=1'" {
		t.Fatalf("restore = %q", restore)
	}
	empty, err := BootArgumentRestoreBuilder("")
	if err != nil || empty != "nvram -d boot-args 2>/dev/null || true" {
		t.Fatalf("empty restore = %q, %v", empty, err)
	}
}

func TestScalarRestoreBuilderValidatesCapturedValue(t *testing.T) {
	builder := ScalarRestoreBuilder("/usr/sbin/sysctl -w debug.lowpri_throttle_enabled=")
	command, err := builder("1")
	if err != nil || command != "/usr/sbin/sysctl -w debug.lowpri_throttle_enabled=1" {
		t.Fatalf("command = %q, error = %v", command, err)
	}
	if _, err := builder("1; touch /tmp/bad"); err == nil {
		t.Fatal("unsafe captured scalar was accepted")
	}
}

func TestSymlinkBuildersRefuseRegularFilesAndRestoreOriginalTarget(t *testing.T) {
	apply := SymlinkApplyBuilder("/usr/local/bin/tool", "/System/tool")
	if _, err := apply("__NON_SYMLINK__"); err == nil {
		t.Fatal("apply builder accepted an existing regular file")
	}
	command, err := apply("/old/tool")
	if err != nil || command != "mkdir -p '/usr/local/bin' && ln -sfn '/System/tool' '/usr/local/bin/tool'" {
		t.Fatalf("apply command = %q, %v", command, err)
	}
	restore := SymlinkRestoreBuilder("/usr/local/bin/tool")
	command, err = restore("/old/tool")
	if err != nil || command != "ln -sfn '/old/tool' '/usr/local/bin/tool'" {
		t.Fatalf("restore command = %q, %v", command, err)
	}
	command, err = restore("")
	if err != nil || command != "rm -f '/usr/local/bin/tool'" {
		t.Fatalf("empty restore command = %q, %v", command, err)
	}
}

func TestPMSetRestoreCommandRestoresEveryPowerProfile(t *testing.T) {
	input := "Battery Power:\n lowpowermode 1\n sleep 5\nAC Power:\n lowpowermode 0\n sleep 20\n"
	command, err := PMSetRestoreCommand(input, "lowpowermode", "sleep")
	if err != nil {
		t.Fatalf("PMSetRestoreCommand: %v", err)
	}
	want := "pmset -b lowpowermode 1 sleep 5 && pmset -c lowpowermode 0 sleep 20"
	if command != want {
		t.Fatalf("command = %q, want %q", command, want)
	}
}

func TestDNSRestoreCommandPreservesOriginalServers(t *testing.T) {
	command := DNSRestoreCommand("Wi-Fi", "9.9.9.9\n149.112.112.112")
	want := "networksetup -setdnsservers 'Wi-Fi' '9.9.9.9' '149.112.112.112'"
	if command != want {
		t.Fatalf("command = %q, want %q", command, want)
	}
	empty := DNSRestoreCommand("Wi-Fi", "There aren't any DNS Servers set on Wi-Fi.")
	if empty != "networksetup -setdnsservers 'Wi-Fi' empty" {
		t.Fatalf("empty command = %q", empty)
	}
}

func TestValueCommandApplySnapshotsBeforeMutation(t *testing.T) {
	oldRunner := commandRunner
	oldSave := saveCommandState
	t.Cleanup(func() {
		commandRunner = oldRunner
		saveCommandState = oldSave
	})
	runner := &queueRunner{outputs: []string{"original", ""}}
	commandRunner = runner
	var savedID, savedState string
	saveCommandState = func(id, state string) error {
		savedID, savedState = id, state
		return nil
	}
	tweak := NewValueCommandTweak("value", "Value", "Value", CategoryOther, RiskLow,
		"echo false", "capture", func(state string) (string, error) { return "apply " + state, nil },
		func(state string) (string, error) { return "restore " + state, nil }, false)

	if err := tweak.Apply(); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if savedID != "value" || savedState != "original" {
		t.Fatalf("saved = %q, %q", savedID, savedState)
	}
	wantCalls := [][]string{{"/bin/sh", "-c", "capture"}, {"/bin/sh", "-c", "apply original"}}
	if !reflect.DeepEqual(runner.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, wantCalls)
	}
}

func TestValueCommandApplyStopsWhenSnapshotFails(t *testing.T) {
	oldRunner := commandRunner
	oldSave := saveCommandState
	t.Cleanup(func() {
		commandRunner = oldRunner
		saveCommandState = oldSave
	})
	runner := &queueRunner{outputs: []string{"original"}}
	commandRunner = runner
	saveCommandState = func(string, string) error { return errors.New("disk full") }
	tweak := NewValueCommandTweak("value", "Value", "Value", CategoryOther, RiskLow,
		"echo false", "capture", func(string) (string, error) { return "apply", nil },
		func(string) (string, error) { return "restore", nil }, false)

	if err := tweak.Apply(); err == nil {
		t.Fatal("Apply returned nil error")
	}
	if len(runner.calls) != 1 {
		t.Fatalf("runner calls = %d, want capture only", len(runner.calls))
	}
}

func TestValueCommandRevertRestoresAndConsumesReceipt(t *testing.T) {
	oldRunner := commandRunner
	oldLoad := loadCommandState
	oldConsume := consumeCommandState
	t.Cleanup(func() {
		commandRunner = oldRunner
		loadCommandState = oldLoad
		consumeCommandState = oldConsume
	})
	runner := &queueRunner{}
	commandRunner = runner
	loadCommandState = func(id string) (string, error) { return "original", nil }
	consumed := false
	consumeCommandState = func(id string) error { consumed = id == "value"; return nil }
	tweak := NewValueCommandTweak("value", "Value", "Value", CategoryOther, RiskLow,
		"echo true", "capture", func(string) (string, error) { return "apply", nil },
		func(state string) (string, error) { return "restore " + state, nil }, false)

	if err := tweak.Revert(); err != nil {
		t.Fatalf("Revert: %v", err)
	}
	if !consumed {
		t.Fatal("receipt was not consumed")
	}
	want := [][]string{{"/bin/sh", "-c", "restore original"}}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, want)
	}
}

func TestActionTweakIsNeverReportedAppliedAndCannotRevert(t *testing.T) {
	tweak := NewActionTweak("flush", "Flush", "Flush", CategoryOther, RiskLow, "true", false)
	if probe := tweak.Probe(); probe.State != ProbeOff || probe.Applied {
		t.Fatalf("Probe = %#v, want off", probe)
	}
	if err := tweak.Revert(); !errors.Is(err, ErrNotReversible) {
		t.Fatalf("Revert error = %v, want ErrNotReversible", err)
	}
}
