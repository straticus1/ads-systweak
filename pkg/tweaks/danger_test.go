package tweaks

import (
	"reflect"
	"testing"
)

func TestVisibleTweaksHidesHighRiskUntilUnlocked(t *testing.T) {
	low := NewCommandTweak("low", "Low", "", CategoryOther, RiskLow, "true", "true", "true", false)
	medium := NewCommandTweak("medium", "Medium", "", CategoryOther, RiskMedium, "true", "true", "true", false)
	high := NewCommandTweak("high", "High", "", CategoryOther, RiskHigh, "true", "true", "true", false)
	registry := []Tweak{low, high, medium}

	if got := VisibleTweaks(registry, false); !reflect.DeepEqual(got, []Tweak{low, medium}) {
		t.Fatalf("locked VisibleTweaks = %#v", got)
	}
	if got := VisibleTweaks(registry, true); !reflect.DeepEqual(got, registry) {
		t.Fatalf("unlocked VisibleTweaks = %#v", got)
	}
	if got := VisibleTweaks(registry, true); len(got) > 0 && &got[0] == &registry[0] {
		t.Fatal("VisibleTweaks returned the registry's backing slice")
	}
}

func TestDangerousTweaksReturnsOnlyHighRiskInRegistryOrder(t *testing.T) {
	highA := NewCommandTweak("high-a", "High A", "", CategoryOther, RiskHigh, "true", "true", "true", false)
	low := NewCommandTweak("low", "Low", "", CategoryOther, RiskLow, "true", "true", "true", false)
	highB := NewCommandTweak("high-b", "High B", "", CategoryOther, RiskHigh, "true", "true", "true", false)

	if got := DangerousTweaks([]Tweak{highA, low, highB}); !reflect.DeepEqual(got, []Tweak{highA, highB}) {
		t.Fatalf("DangerousTweaks = %#v", got)
	}
}

func TestValidateDangerUnlockRequiresAcknowledgementAndExactPhrase(t *testing.T) {
	for _, test := range []struct {
		acknowledged bool
		phrase       string
		want         bool
	}{
		{true, DangerUnlockPhrase, true},
		{false, DangerUnlockPhrase, false},
		{true, "I know what I am doing", false},
		{true, " " + DangerUnlockPhrase, false},
		{true, DangerUnlockPhrase + " ", false},
	} {
		if got := ValidateDangerUnlock(test.acknowledged, test.phrase); got != test.want {
			t.Fatalf("ValidateDangerUnlock(%v, %q) = %v, want %v", test.acknowledged, test.phrase, got, test.want)
		}
	}
}
