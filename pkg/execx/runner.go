// Package execx provides the process-execution boundary used by system tools.
package execx

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Runner executes a program directly with an argument vector.
type Runner interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
}

// OSRunner executes programs using os/exec without invoking a shell.
type OSRunner struct{}

// CommandError describes a failed program and preserves its stderr output.
type CommandError struct {
	Name   string
	Args   []string
	Stderr string
	Err    error
}

func (e *CommandError) Error() string {
	return fmt.Sprintf("%s failed: %v; stderr: %s", e.Name, e.Err, e.Stderr)
}

func (e *CommandError) Unwrap() error { return e.Err }

// Run executes name directly and returns stdout with surrounding whitespace removed.
func (OSRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", &CommandError{
			Name:   name,
			Args:   append([]string(nil), args...),
			Stderr: strings.TrimSpace(stderr.String()),
			Err:    err,
		}
	}
	return strings.TrimSpace(stdout.String()), nil
}

// AdministratorScript creates an AppleScript command without adding another shell layer.
func AdministratorScript(command string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`"`, `\"`,
		"\r", `\r`,
		"\n", `\n`,
	)
	return `do shell script "` + replacer.Replace(command) + `" with administrator privileges`
}
