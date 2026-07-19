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
	out, err := RunShell(fmt.Sprintf("defaults read %s %s", d.domain, d.key))
	if err != nil {
		return false, nil // Assume not applied if key is missing
	}
	out = strings.TrimSpace(out)

	if d.valType == "bool" {
		expected := "0"
		if v, ok := d.targetVal.(bool); ok {
			if v {
				expected = "1"
			}
		} else {
			return false, fmt.Errorf("internal type error: expected bool for %s", d.ID())
		}

		return out == expected || out == "true" || out == fmt.Sprintf("%v", d.targetVal), nil
	}

	return out == fmt.Sprintf("%v", d.targetVal), nil
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
	out, err := RunShell(c.checkCmd)
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(out) == "true" || strings.TrimSpace(out) == "1" || strings.TrimSpace(out) == "yes", nil
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
