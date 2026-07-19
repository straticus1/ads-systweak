package tweaks

import (
	"errors"
	"testing"
)

type stubTweak struct {
	id    string
	probe ProbeResult
}

func (s stubTweak) ID() string               { return s.id }
func (s stubTweak) Name() string             { return s.id }
func (s stubTweak) Description() string      { return s.id }
func (s stubTweak) Category() TweakCategory  { return CategoryOther }
func (s stubTweak) RiskLevel() RiskLevel     { return RiskLow }
func (s stubTweak) Probe() ProbeResult       { return s.probe }
func (s stubTweak) IsApplied() (bool, error) { return s.probe.Applied, s.probe.Err }
func (s stubTweak) Apply() error             { return nil }
func (s stubTweak) Revert() error            { return nil }

func TestBuildPlanCreatesApplyAndRevertActions(t *testing.T) {
	registry := []Tweak{
		stubTweak{id: "off", probe: ProbeResult{State: ProbeOff}},
		stubTweak{id: "on", probe: ProbeResult{State: ProbeApplied, Applied: true}},
	}
	desired := map[string]bool{"off": true, "on": false}

	plan := BuildPlan(registry, desired)
	if len(plan) != 2 {
		t.Fatalf("plan length = %d, want 2", len(plan))
	}
	if plan[0].Action != PlanApply || plan[1].Action != PlanRevert {
		t.Fatalf("actions = %q, %q", plan[0].Action, plan[1].Action)
	}
}

func TestBuildPlanBlocksUnknownState(t *testing.T) {
	probeErr := errors.New("permission denied")
	registry := []Tweak{stubTweak{id: "unknown", probe: ProbeResult{State: ProbePermissionDenied, Err: probeErr}}}

	plan := BuildPlan(registry, map[string]bool{"unknown": true})
	if len(plan) != 1 {
		t.Fatalf("plan length = %d, want 1", len(plan))
	}
	if plan[0].Action != PlanBlocked || !errors.Is(plan[0].Probe.Err, probeErr) {
		t.Fatalf("plan item = %#v, want blocked permission error", plan[0])
	}
}

func TestBuildPlanOmitsMatchingAndUnconfiguredTweaks(t *testing.T) {
	registry := []Tweak{
		stubTweak{id: "matching", probe: ProbeResult{State: ProbeApplied, Applied: true}},
		stubTweak{id: "unset", probe: ProbeResult{State: ProbeOff}},
	}

	plan := BuildPlan(registry, map[string]bool{"matching": true})
	if len(plan) != 0 {
		t.Fatalf("plan = %#v, want empty", plan)
	}
}
