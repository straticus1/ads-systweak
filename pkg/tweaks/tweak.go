package tweaks

// TweakCategory groups tweaks in the UI and CLI
type TweakCategory string

const (
	CategoryDisk           TweakCategory = "Disk"
	CategorySystem         TweakCategory = "System"
	CategoryNetwork        TweakCategory = "Network"
	CategoryNetworkStorage TweakCategory = "Network Storage"
	CategoryLowLevel       TweakCategory = "Low Level"
	CategoryMemory         TweakCategory = "Memory"
	CategoryKernel         TweakCategory = "Kernel"
	CategoryApps           TweakCategory = "Apps"
	CategoryHiddenCLI      TweakCategory = "Hidden CLI Tools"
	CategoryOther          TweakCategory = "Other"
)

type RiskLevel string

const (
	RiskLow    RiskLevel = "Low"
	RiskMedium RiskLevel = "Medium"
	RiskHigh   RiskLevel = "High"
)

// Tweak represents a single configurable setting
type Tweak interface {
	ID() string          // Unique identifier, used in CLI (e.g., "show-hidden-files")
	Name() string        // Human-readable name
	Description() string // Detailed explanation
	Category() TweakCategory
	RiskLevel() RiskLevel

	IsApplied() (bool, error) // Returns true if the tweak is currently active
	Apply() error             // Enables the tweak
	Revert() error            // Disables the tweak
}

var Registry []Tweak
var DryRun bool
var Verbose bool

// Register adds a tweak to the global registry
func Register(t Tweak) {
	Registry = append(Registry, t)
}

// GetByCategory returns all tweaks belonging to a specific category
func GetByCategory(cat TweakCategory) []Tweak {
	var res []Tweak
	for _, t := range Registry {
		if t.Category() == cat {
			res = append(res, t)
		}
	}
	return res
}
