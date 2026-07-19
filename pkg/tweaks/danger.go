package tweaks

// DangerUnlockPhrase is the exact acknowledgement required to reveal High-risk tweaks.
const DangerUnlockPhrase = "I KNOW WHAT I AM DOING"

// ValidateDangerUnlock accepts only an explicit acknowledgement and exact phrase.
func ValidateDangerUnlock(acknowledged bool, phrase string) bool {
	return acknowledged && phrase == DangerUnlockPhrase
}

// VisibleTweaks returns the registry entries visible at the current unlock level.
func VisibleTweaks(registry []Tweak, dangerUnlocked bool) []Tweak {
	visible := make([]Tweak, 0, len(registry))
	for _, tweak := range registry {
		if dangerUnlocked || tweak.RiskLevel() != RiskHigh {
			visible = append(visible, tweak)
		}
	}
	return visible
}

// DangerousTweaks returns only High-risk registry entries.
func DangerousTweaks(registry []Tweak) []Tweak {
	dangerous := make([]Tweak, 0)
	for _, tweak := range registry {
		if tweak.RiskLevel() == RiskHigh {
			dangerous = append(dangerous, tweak)
		}
	}
	return dangerous
}
