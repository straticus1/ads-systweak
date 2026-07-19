package tweaks

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var ErrNotReversible = errors.New("one-shot action is not reversible")

// CommandBuilder turns a captured pre-mutation value into an apply or restore command.
type CommandBuilder func(captured string) (string, error)

type commandStateReceipt struct {
	ID        string `json:"id"`
	State     string `json:"state"`
	Timestamp string `json:"timestamp"`
}

var (
	saveCommandState    = saveCommandStateToDisk
	loadCommandState    = loadCommandStateFromDisk
	consumeCommandState = consumeCommandStateFromDisk
)

// ValueCommandTweak snapshots command state before applying and restores that exact state.
type ValueCommandTweak struct {
	*CommandTweak
	captureCmd     string
	applyBuilder   CommandBuilder
	restoreBuilder CommandBuilder
}

// ActionTweak represents a one-shot operation rather than persistent state.
type ActionTweak struct{ *CommandTweak }

func NewActionTweak(id, name, desc string, cat TweakCategory, risk RiskLevel, command string, root bool) *ActionTweak {
	return &ActionTweak{CommandTweak: NewCommandTweak(id, name, desc, cat, risk, `echo "false"`, command, "", root)}
}

func (a *ActionTweak) Probe() ProbeResult       { return ProbeResult{State: ProbeOff} }
func (a *ActionTweak) IsApplied() (bool, error) { return false, nil }
func (a *ActionTweak) Revert() error            { return ErrNotReversible }

func NewValueCommandTweak(id, name, desc string, cat TweakCategory, risk RiskLevel, check, capture string, apply, restore CommandBuilder, root bool) *ValueCommandTweak {
	return &ValueCommandTweak{
		CommandTweak:   NewCommandTweak(id, name, desc, cat, risk, check, "", "", root),
		captureCmd:     capture,
		applyBuilder:   apply,
		restoreBuilder: restore,
	}
}

func (v *ValueCommandTweak) Apply() error {
	if DryRun {
		fmt.Printf("[DRY RUN] Capture and apply ValueCommandTweak (%s)\n", v.id)
		return nil
	}
	captured, err := RunShell(v.captureCmd)
	if err != nil {
		return fmt.Errorf("capture %s state: %w", v.id, err)
	}
	if err := saveCommandState(v.id, captured); err != nil {
		return fmt.Errorf("save %s state: %w", v.id, err)
	}
	command, err := v.applyBuilder(captured)
	if err != nil {
		return fmt.Errorf("build %s apply command: %w", v.id, err)
	}
	return v.run(command)
}

func (v *ValueCommandTweak) Revert() error {
	if DryRun {
		fmt.Printf("[DRY RUN] Restore ValueCommandTweak (%s) from captured state\n", v.id)
		return nil
	}
	captured, err := loadCommandState(v.id)
	if err != nil {
		return fmt.Errorf("load %s state: %w", v.id, err)
	}
	command, err := v.restoreBuilder(captured)
	if err != nil {
		return fmt.Errorf("build %s restore command: %w", v.id, err)
	}
	if err := v.run(command); err != nil {
		return err
	}
	return consumeCommandState(v.id)
}

func (v *ValueCommandTweak) run(command string) error {
	if strings.TrimSpace(command) == "" {
		return errors.New("refusing to execute an empty command")
	}
	if v.runAsRoot {
		_, err := RunPrivileged(command)
		return err
	}
	_, err := RunShell(command)
	return err
}

// AddBootArgument adds one NVRAM token without disturbing existing tokens.
func AddBootArgument(current, argument string) string {
	fields := strings.Fields(current)
	for _, field := range fields {
		if field == argument {
			return strings.Join(fields, " ")
		}
	}
	return strings.TrimSpace(strings.Join(append(fields, argument), " "))
}

// RemoveBootArgument removes one NVRAM token without disturbing the others.
func RemoveBootArgument(current, argument string) string {
	fields := strings.Fields(current)
	result := fields[:0]
	for _, field := range fields {
		if field != argument {
			result = append(result, field)
		}
	}
	return strings.Join(result, " ")
}

// ParseNVRAMBootArgs removes nvram's key prefix from its output.
func ParseNVRAMBootArgs(output string) string {
	trimmed := strings.TrimSpace(output)
	if index := strings.IndexAny(trimmed, "\t "); index >= 0 && strings.TrimSpace(trimmed[:index]) == "boot-args" {
		return strings.TrimSpace(trimmed[index+1:])
	}
	return trimmed
}

// BootArgumentApplyBuilder adds one token while preserving all captured boot arguments.
func BootArgumentApplyBuilder(argument string) CommandBuilder {
	return func(captured string) (string, error) {
		value := AddBootArgument(ParseNVRAMBootArgs(captured), argument)
		if value == "" {
			return "", errors.New("boot argument cannot be empty")
		}
		return "nvram boot-args=" + shellQuote(value), nil
	}
}

// BootArgumentRestoreBuilder restores the exact captured boot-args value.
func BootArgumentRestoreBuilder(captured string) (string, error) {
	value := ParseNVRAMBootArgs(captured)
	if value == "" {
		return "nvram -d boot-args 2>/dev/null || true", nil
	}
	return "nvram boot-args=" + shellQuote(value), nil
}

// StaticCommandBuilder returns the same trusted built-in command for every snapshot.
func StaticCommandBuilder(command string) CommandBuilder {
	return func(string) (string, error) { return command, nil }
}

// ScalarRestoreBuilder appends one validated scalar captured from a system command.
func ScalarRestoreBuilder(prefix string) CommandBuilder {
	return func(captured string) (string, error) {
		value := strings.TrimSpace(captured)
		if !safeShellToken(value) {
			return "", errors.New("captured scalar contains unsafe characters")
		}
		return prefix + value, nil
	}
}

// DNSRestoreCommand restores the exact ordered DNS server list for a service.
func DNSRestoreCommand(service, captured string) string {
	servers := strings.Fields(captured)
	if strings.Contains(strings.ToLower(captured), "aren't any dns servers") || len(servers) == 0 {
		return "networksetup -setdnsservers " + shellQuote(service) + " empty"
	}
	quoted := make([]string, 0, len(servers))
	for _, server := range servers {
		quoted = append(quoted, shellQuote(server))
	}
	return "networksetup -setdnsservers " + shellQuote(service) + " " + strings.Join(quoted, " ")
}

func DNSRestoreBuilder(service string) CommandBuilder {
	return func(captured string) (string, error) { return DNSRestoreCommand(service, captured), nil }
}

func PMSetRestoreBuilder(keys ...string) CommandBuilder {
	return func(captured string) (string, error) { return PMSetRestoreCommand(captured, keys...) }
}

// PMSetRestoreCommand restores selected settings for every power profile present.
func PMSetRestoreCommand(captured string, keys ...string) (string, error) {
	profiles := parsePMSet(captured)
	type profile struct {
		name string
		flag string
	}
	ordered := []profile{{"Battery Power", "-b"}, {"AC Power", "-c"}, {"UPS Power", "-u"}}
	var commands []string
	for _, item := range ordered {
		values := profiles[item.name]
		if len(values) == 0 {
			continue
		}
		parts := []string{"pmset", item.flag}
		for _, key := range keys {
			value, ok := values[key]
			if !ok {
				continue
			}
			if !safeShellToken(key) || !safeShellToken(value) {
				return "", errors.New("captured pmset value contains unsafe characters")
			}
			parts = append(parts, key, value)
		}
		if len(parts) > 2 {
			commands = append(commands, strings.Join(parts, " "))
		}
	}
	if len(commands) == 0 {
		return "", errors.New("captured pmset output did not contain requested keys")
	}
	return strings.Join(commands, " && "), nil
}

func safeShellToken(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || strings.ContainsRune("._-", char) {
			continue
		}
		return false
	}
	return true
}

func parsePMSet(output string) map[string]map[string]string {
	profiles := make(map[string]map[string]string)
	current := ""
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasSuffix(trimmed, ":") {
			current = strings.TrimSuffix(trimmed, ":")
			profiles[current] = make(map[string]string)
			continue
		}
		fields := strings.Fields(trimmed)
		if current != "" && len(fields) >= 2 {
			profiles[current][fields[0]] = fields[1]
		}
	}
	return profiles
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func commandStatePath(id string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(id))
	return filepath.Join(home, ".ads-systweak-backups", fmt.Sprintf("command-%x.json", hash)), nil
}

func saveCommandStateToDisk(id, state string) error {
	path, err := commandStatePath(id)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	receipt := commandStateReceipt{ID: id, State: state, Timestamp: time.Now().Format(time.RFC3339Nano)}
	data, err := json.Marshal(receipt)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".command-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func loadCommandStateFromDisk(id string) (string, error) {
	path, err := commandStatePath(id)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var receipt commandStateReceipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		return "", err
	}
	if receipt.ID != id {
		return "", errors.New("command state receipt identity mismatch")
	}
	return receipt.State, nil
}

func consumeCommandStateFromDisk(id string) error {
	path, err := commandStatePath(id)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

// stableMapKeys is retained for deterministic command builders added by registry adapters.
func stableMapKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
