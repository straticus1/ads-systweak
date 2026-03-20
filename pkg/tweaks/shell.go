package tweaks

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// RunShell executes an arbitrary shell command
func RunShell(command string) (string, error) {
	if Verbose {
		fmt.Printf("[EXEC] %s\n", command)
	}
	cmd := exec.Command("sh", "-c", command)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("command failed: %s, stderr: %s, err: %w", command, stderr.String(), err)
	}
	return strings.TrimSpace(out.String()), nil
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
	escapedCmd := strings.ReplaceAll(command, "\"", "\\\"")
	fullCmd := fmt.Sprintf(`osascript -e 'do shell script "%s" with administrator privileges'`, escapedCmd)
	return RunShell(fullCmd)
}
