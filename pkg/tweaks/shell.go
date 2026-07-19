package tweaks

import (
	"context"
	"fmt"
	"strings"

	"ads-systweak/pkg/execx"
)

var commandRunner execx.Runner = execx.OSRunner{}

// RunShell executes an arbitrary shell command
func RunShell(command string) (string, error) {
	if Verbose {
		fmt.Printf("[EXEC] %s\n", command)
	}
	return commandRunner.Run(context.Background(), "/bin/sh", "-c", command)
}

// RunCommand executes a program directly. Use this for all user-derived arguments.
func RunCommand(name string, args ...string) (string, error) {
	if Verbose {
		fmt.Printf("[EXEC] %s %s\n", name, strings.Join(args, " "))
	}
	return commandRunner.Run(context.Background(), name, args...)
}

// CheckDefaultsBool checks if a boolean default is set to true
func CheckDefaultsBool(domain, key string) (bool, error) {
	out, err := RunShell(fmt.Sprintf("defaults read %s %s", domain, key))
	if err != nil {
		// defaults read fails if key does not exist
		return false, nil
	}
	out = strings.TrimSpace(out)
	if out == "1" || out == "true" {
		return true, nil
	}
	return false, nil
}

// CheckDefaultsString reads a string default
func CheckDefaultsString(domain, key string) (string, error) {
	out, err := RunShell(fmt.Sprintf("defaults read %s %s", domain, key))
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(out), nil
}

// WriteDefaultsBool writes a boolean to a defaults domain
func WriteDefaultsBool(domain, key string, val bool) error {
	v := "false"
	if val {
		v = "true"
	}
	_, err := RunShell(fmt.Sprintf("defaults write %s %s -bool %s", domain, key, v))
	return err
}

// WriteDefaultsString writes a string to a defaults domain
func WriteDefaultsString(domain, key, val string) error {
	_, err := RunShell(fmt.Sprintf("defaults write %s %s -string \"%s\"", domain, key, val))
	return err
}

// DeleteDefaultsKey removes a key from a defaults domain
func DeleteDefaultsKey(domain, key string) error {
	_, err := RunShell(fmt.Sprintf("defaults delete %s %s", domain, key))
	return err
}

// RestartApp gracefully restarts a macOS app (like Finder or Dock)
func RestartApp(appName string) error {
	_, err := RunShell(fmt.Sprintf("killall %s", appName))
	return err
}

// RunPrivileged runs a script via osascript to prompt for admin privileges
func RunPrivileged(command string) (string, error) {
	return RunCommand("/usr/bin/osascript", "-e", execx.AdministratorScript(command))
}
