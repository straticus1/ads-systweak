package execx

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestOSRunnerPassesArgumentsWithoutShellInterpretation(t *testing.T) {
	t.Parallel()

	out, err := (OSRunner{}).Run(context.Background(), "/usr/bin/printf", "%s", "safe; echo injected")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if out != "safe; echo injected" {
		t.Fatalf("Run output = %q, want literal argument", out)
	}
}

func TestOSRunnerReturnsStructuredStderr(t *testing.T) {
	t.Parallel()

	_, err := (OSRunner{}).Run(context.Background(), "/bin/sh", "-c", "echo broken >&2; exit 7")
	if err == nil {
		t.Fatal("Run returned nil error")
	}

	var commandErr *CommandError
	if !errors.As(err, &commandErr) {
		t.Fatalf("error type = %T, want *CommandError", err)
	}
	if commandErr.Stderr != "broken" {
		t.Fatalf("stderr = %q, want %q", commandErr.Stderr, "broken")
	}
	if !strings.Contains(err.Error(), "exit status 7") {
		t.Fatalf("error %q does not contain exit status", err)
	}
}

func TestAdministratorScriptEscapesAppleScriptString(t *testing.T) {
	t.Parallel()

	script := AdministratorScript("printf \"quoted\"; printf '\\path'\nnext")
	want := `do shell script "printf \"quoted\"; printf '\\path'\nnext" with administrator privileges`
	if script != want {
		t.Fatalf("AdministratorScript() = %q, want %q", script, want)
	}
}
