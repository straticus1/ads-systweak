package tweaks

import (
	"errors"
	"fmt"
	"strings"

	"ads-systweak/pkg/backup"
)

var (
	saveDefaultsBackup    = backup.SaveBackup
	restoreDefaultsBackup = backup.Restore
)

type DefaultsTweak struct {
	id         string
	name       string
	desc       string
	cat        TweakCategory
	risk       RiskLevel
	domain     string
	key        string
	valType    string // bool, string, int, float
	targetVal  interface{}
	restartApp string
}

func NewDefaultsTweak(id, name, desc string, cat TweakCategory, risk RiskLevel, domain, key, valType string, targetVal interface{}, restart string) *DefaultsTweak {
	return &DefaultsTweak{id, name, desc, cat, risk, domain, key, valType, targetVal, restart}
}

func (d *DefaultsTweak) ID() string              { return d.id }
func (d *DefaultsTweak) Name() string            { return d.name }
func (d *DefaultsTweak) Description() string     { return d.desc }
func (d *DefaultsTweak) Category() TweakCategory { return d.cat }
func (d *DefaultsTweak) RiskLevel() RiskLevel    { return d.risk }

func (d *DefaultsTweak) IsApplied() (bool, error) {
	result := d.Probe()
	return result.Applied, result.Err
}

func (d *DefaultsTweak) Probe() ProbeResult {
	out, err := RunCommand("/usr/bin/defaults", "read", d.domain, d.key)
	if err != nil {
		return classifyProbeError(err, true)
	}
	out = strings.TrimSpace(out)

	if d.valType == "bool" {
		target, ok := d.targetVal.(bool)
		if !ok {
			err := fmt.Errorf("internal type error: expected bool for %s", d.ID())
			return ProbeResult{State: ProbeError, Err: err}
		}
		actual, ok := parseDefaultsBool(out)
		if !ok {
			err := fmt.Errorf("unexpected boolean value %q for %s", out, d.ID())
			return ProbeResult{State: ProbeError, Err: err}
		}
		return appliedProbe(actual == target)
	}

	return appliedProbe(out == fmt.Sprintf("%v", d.targetVal))
}

func (d *DefaultsTweak) Apply() error {
	var valStr string
	if d.valType == "bool" {
		if v, ok := d.targetVal.(bool); ok {
			valStr = "false"
			if v {
				valStr = "true"
			}
		} else {
			return fmt.Errorf("internal type error: expected bool for %s", d.ID())
		}
	} else {
		valStr = fmt.Sprintf("%v", d.targetVal)
	}

	// For floats/ints we might need specific types, but generally string/bool covers most
	typeFlag := ""
	switch d.valType {
	case "bool":
		typeFlag = "-bool"
	case "int":
		typeFlag = "-int"
	case "float":
		typeFlag = "-float"
	case "string":
		typeFlag = "-string"
		valStr = fmt.Sprintf("\"%s\"", valStr)
	}

	cmd := fmt.Sprintf("defaults write %s %s %s %s", d.domain, d.key, typeFlag, valStr)

	if DryRun {
		fmt.Printf("[DRY RUN] Apply DefaultsTweak (%s): %s\n", d.id, cmd)
		return nil
	}

	if err := saveDefaultsBackup(d.domain, d.key); err != nil {
		return fmt.Errorf("back up %s/%s: %w", d.domain, d.key, err)
	}

	_, err := RunShell(cmd)
	if err != nil {
		return err
	}

	if d.restartApp != "" {
		if err := RestartApp(d.restartApp); err != nil {
			return err
		}
	}
	return nil
}

func (d *DefaultsTweak) Revert() error {
	if DryRun {
		fmt.Printf("[DRY RUN] Restore DefaultsTweak (%s) from backup\n", d.id)
		return nil
	}

	err := restoreDefaultsBackup(d.domain, d.key)
	if errors.Is(err, backup.ErrNoBackup) {
		_, err = RunCommand("/usr/bin/defaults", "delete", d.domain, d.key)
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "does not exist") {
			return err
		}
	} else if err != nil {
		return err
	}
	if d.restartApp != "" {
		if err := RestartApp(d.restartApp); err != nil {
			return err
		}
	}
	return nil
}

type CommandTweak struct {
	id        string
	name      string
	desc      string
	cat       TweakCategory
	risk      RiskLevel
	checkCmd  string
	applyCmd  string
	revertCmd string
	runAsRoot bool
}

func NewCommandTweak(id, name, desc string, cat TweakCategory, risk RiskLevel, check, apply, revert string, root bool) *CommandTweak {
	return &CommandTweak{id, name, desc, cat, risk, check, apply, revert, root}
}

func (c *CommandTweak) ID() string              { return c.id }
func (c *CommandTweak) Name() string            { return c.name }
func (c *CommandTweak) Description() string     { return c.desc }
func (c *CommandTweak) Category() TweakCategory { return c.cat }
func (c *CommandTweak) RiskLevel() RiskLevel    { return c.risk }

func (c *CommandTweak) IsApplied() (bool, error) {
	result := c.Probe()
	return result.Applied, result.Err
}

func (c *CommandTweak) Probe() ProbeResult {
	out, err := RunShell(c.checkCmd)
	if err != nil {
		return classifyProbeError(err, false)
	}
	normalized := strings.ToLower(strings.TrimSpace(out))
	return appliedProbe(normalized == "true" || normalized == "1" || normalized == "yes")
}

func parseDefaultsBool(value string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes":
		return true, true
	case "0", "false", "no":
		return false, true
	default:
		return false, false
	}
}

func appliedProbe(applied bool) ProbeResult {
	if applied {
		return ProbeResult{State: ProbeApplied, Applied: true}
	}
	return ProbeResult{State: ProbeOff}
}

func classifyProbeError(err error, missingIsOff bool) ProbeResult {
	text := strings.ToLower(err.Error())
	if missingIsOff && (strings.Contains(text, "does not exist") || strings.Contains(text, "not found")) {
		return ProbeResult{State: ProbeOff}
	}
	if strings.Contains(text, "permission denied") || strings.Contains(text, "not permitted") || strings.Contains(text, "not authorized") {
		return ProbeResult{State: ProbePermissionDenied, Err: err}
	}
	if strings.Contains(text, "executable file not found") || strings.Contains(text, "no such file or directory") {
		return ProbeResult{State: ProbeUnsupported, Err: err}
	}
	return ProbeResult{State: ProbeError, Err: err}
}

func (c *CommandTweak) Apply() error {
	if DryRun {
		fmt.Printf("[DRY RUN] Apply CommandTweak (%s): %s (Root: %v)\n", c.id, c.applyCmd, c.runAsRoot)
		return nil
	}
	if c.runAsRoot {
		_, err := RunPrivileged(c.applyCmd)
		return err
	}
	_, err := RunShell(c.applyCmd)
	return err
}

func (c *CommandTweak) Revert() error {
	if DryRun {
		fmt.Printf("[DRY RUN] Revert CommandTweak (%s): %s (Root: %v)\n", c.id, c.revertCmd, c.runAsRoot)
		return nil
	}
	if c.runAsRoot {
		_, err := RunPrivileged(c.revertCmd)
		return err
	}
	_, err := RunShell(c.revertCmd)
	return err
}
