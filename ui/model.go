package ui

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"ads-systweak/pkg/tweaks"
	"howett.net/plist"
)

type DefaultRow struct {
	Key   string
	Type  string
	Value interface{}
}

type PlanSummary struct {
	Apply    int
	Revert   int
	Blocked  int
	HighRisk int
}

// DangerZoneState is intentionally process-local. It has no serialization path.
type DangerZoneState struct {
	unlocked bool
}

func NewDangerZoneState() *DangerZoneState { return &DangerZoneState{} }

func (s *DangerZoneState) Unlocked() bool { return s.unlocked }

func (s *DangerZoneState) Unlock(acknowledged bool, phrase string) bool {
	if tweaks.ValidateDangerUnlock(acknowledged, phrase) {
		s.unlocked = true
	}
	return s.unlocked
}

func (s *DangerZoneState) VisibleTweaks(registry []tweaks.Tweak) []tweaks.Tweak {
	return tweaks.VisibleTweaks(registry, s.unlocked)
}

func TweakCategories() []tweaks.TweakCategory {
	return []tweaks.TweakCategory{
		tweaks.CategorySystem,
		tweaks.CategoryDisk,
		tweaks.CategoryNetwork,
		tweaks.CategoryNetworkStorage,
		tweaks.CategoryApps,
		tweaks.CategoryLowLevel,
		tweaks.CategoryMemory,
		tweaks.CategoryKernel,
		tweaks.CategoryHiddenCLI,
		tweaks.CategoryOther,
	}
}

func FilterDomains(domains []string, query string) []string {
	query = strings.ToLower(strings.TrimSpace(query))
	result := make([]string, 0, len(domains))
	for _, domain := range domains {
		if query == "" || strings.Contains(strings.ToLower(domain), query) {
			result = append(result, domain)
		}
	}
	sort.Slice(result, func(i, j int) bool { return strings.ToLower(result[i]) < strings.ToLower(result[j]) })
	return result
}

func DecodeDefaultsExport(data []byte) ([]DefaultRow, error) {
	values := make(map[string]interface{})
	if err := plist.NewDecoder(bytes.NewReader(data)).Decode(&values); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	rows := make([]DefaultRow, 0, len(keys))
	for _, key := range keys {
		rows = append(rows, DefaultRow{Key: key, Type: defaultsValueType(values[key]), Value: values[key]})
	}
	return rows, nil
}

func defaultsValueType(value interface{}) string {
	switch value.(type) {
	case bool:
		return "bool"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "int"
	case float32, float64:
		return "float"
	case string:
		return "string"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "dict"
	default:
		return fmt.Sprintf("%T", value)
	}
}

func ParseScalarInput(kind, input string) (string, string, error) {
	switch strings.ToLower(kind) {
	case "bool":
		value, err := strconv.ParseBool(strings.TrimSpace(input))
		if err != nil {
			return "", "", errors.New("boolean must be true or false")
		}
		return "-bool", strconv.FormatBool(value), nil
	case "int":
		value, err := strconv.ParseInt(strings.TrimSpace(input), 10, 64)
		if err != nil {
			return "", "", errors.New("integer is invalid or out of range")
		}
		return "-int", strconv.FormatInt(value, 10), nil
	case "float":
		value, err := strconv.ParseFloat(strings.TrimSpace(input), 64)
		if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
			return "", "", errors.New("float is invalid or out of range")
		}
		return "-float", strconv.FormatFloat(value, 'g', -1, 64), nil
	case "string":
		if strings.ContainsRune(input, 0) {
			return "", "", errors.New("string contains NUL")
		}
		return "-string", input, nil
	default:
		return "", "", fmt.Errorf("editing %s values is not supported", kind)
	}
}

func SummarizePlan(plan []tweaks.PlanItem) PlanSummary {
	var summary PlanSummary
	for _, item := range plan {
		switch item.Action {
		case tweaks.PlanApply:
			summary.Apply++
		case tweaks.PlanRevert:
			summary.Revert++
		case tweaks.PlanBlocked:
			summary.Blocked++
		}
		if item.Action != tweaks.PlanBlocked && item.Tweak.RiskLevel() == tweaks.RiskHigh {
			summary.HighRisk++
		}
	}
	return summary
}

// CanExecutePlan prevents a visual interface from applying High-risk work that is
// hidden behind its session-only Danger Zone gate.
func CanExecutePlan(plan []tweaks.PlanItem, dangerUnlocked bool) error {
	if dangerUnlocked {
		return nil
	}
	for _, item := range plan {
		if item.Action != tweaks.PlanBlocked && item.Tweak.RiskLevel() == tweaks.RiskHigh {
			return errors.New("unlock Danger Zone before applying staged High-risk changes")
		}
	}
	return nil
}
