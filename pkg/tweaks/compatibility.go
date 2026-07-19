package tweaks

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var ErrRequirementMissing = errors.New("required macOS capability is missing")

type RequiredTweak struct {
	Tweak
	requirements []string
}

func RequirePaths(tweak Tweak, requirements ...string) Tweak {
	return &RequiredTweak{Tweak: tweak, requirements: append([]string(nil), requirements...)}
}

func (r *RequiredTweak) Probe() ProbeResult {
	for _, requirement := range r.requirements {
		if err := checkRequirement(requirement); err != nil {
			return ProbeResult{State: ProbeUnsupported, Err: err}
		}
	}
	return r.Tweak.Probe()
}

func (r *RequiredTweak) IsApplied() (bool, error) {
	probe := r.Probe()
	return probe.Applied, probe.Err
}

func checkRequirement(requirement string) error {
	if strings.ContainsRune(requirement, os.PathSeparator) {
		info, err := os.Stat(requirement)
		if err != nil || info.IsDir() {
			return fmt.Errorf("%w: %s", ErrRequirementMissing, requirement)
		}
		return nil
	}
	if _, err := exec.LookPath(requirement); err != nil {
		return fmt.Errorf("%w: %s", ErrRequirementMissing, requirement)
	}
	return nil
}
