package tweaks

import (
	"testing"
)

func TestRegistryCompleteness(t *testing.T) {
	if len(Registry) == 0 {
		t.Fatal("Registry should not be empty after init()")
	}

	seenIDs := make(map[string]bool)

	for _, tt := range Registry {
		if tt.ID() == "" {
			t.Errorf("Found tweak with empty ID")
		}
		if tt.Name() == "" {
			t.Errorf("Tweak '%s' has empty name", tt.ID())
		}
		if tt.Description() == "" {
			t.Errorf("Tweak '%s' has empty description", tt.ID())
		}
		
		if string(tt.RiskLevel()) == "" {
			t.Errorf("Tweak '%s' has empty risk level", tt.ID())
		}

		if seenIDs[tt.ID()] {
			t.Errorf("Duplicate Tweak ID found: %s", tt.ID())
		}
		seenIDs[tt.ID()] = true
	}
}

func TestPresets(t *testing.T) {
	if len(Presets) == 0 {
		t.Fatal("Presets list should not be empty")
	}

	for _, p := range Presets {
		if p.Name == "" {
			t.Errorf("Found preset with empty name")
		}
		if len(p.TweakIDs) == 0 {
			t.Errorf("Preset '%s' has no tweaks", p.Name)
		}
		
		for _, tweakID := range p.TweakIDs {
			found := false
			for _, regTweak := range Registry {
				if regTweak.ID() == tweakID {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Preset '%s' references unknown tweak ID: '%s'", p.Name, tweakID)
			}
		}
	}
}
