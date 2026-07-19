package tweaks

// PlanAction is the operation required to reach desired state.
type PlanAction string

const (
	PlanApply   PlanAction = "apply"
	PlanRevert  PlanAction = "revert"
	PlanBlocked PlanAction = "blocked"
)

// PlanItem is one required or blocked change.
type PlanItem struct {
	Tweak   Tweak
	Desired bool
	Probe   ProbeResult
	Action  PlanAction
}

// BuildPlan compares desired state with observed state and retains blocked probes.
func BuildPlan(registry []Tweak, desired map[string]bool) []PlanItem {
	var plan []PlanItem
	for _, tweak := range registry {
		want, configured := desired[tweak.ID()]
		if !configured {
			continue
		}
		probe := tweak.Probe()
		if probe.State != ProbeApplied && probe.State != ProbeOff {
			plan = append(plan, PlanItem{Tweak: tweak, Desired: want, Probe: probe, Action: PlanBlocked})
			continue
		}
		if want == probe.Applied {
			continue
		}
		action := PlanApply
		if !want {
			action = PlanRevert
		}
		plan = append(plan, PlanItem{Tweak: tweak, Desired: want, Probe: probe, Action: action})
	}
	return plan
}
