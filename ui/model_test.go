package ui

import (
	"reflect"
	"testing"

	"ads-systweak/pkg/tweaks"
	"howett.net/plist"
)

func TestTweakCategoriesIncludesHiddenCLI(t *testing.T) {
	categories := TweakCategories()
	found := false
	for _, category := range categories {
		if category == tweaks.CategoryHiddenCLI {
			found = true
		}
	}
	if !found {
		t.Fatalf("categories = %#v, missing Hidden CLI", categories)
	}
}

func TestFilterDomainsIsCaseInsensitiveAndSorted(t *testing.T) {
	got := FilterDomains([]string{"com.Zeta", "NSGlobalDomain", "com.apple.finder"}, "COM.")
	want := []string{"com.apple.finder", "com.Zeta"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FilterDomains = %#v, want %#v", got, want)
	}
}

func TestDecodeDefaultsExportReturnsRealTypedRows(t *testing.T) {
	data, err := plist.Marshal(map[string]interface{}{
		"Enabled": true,
		"Count":   int64(3),
		"Name":    "Mac",
		"Items":   []interface{}{"one"},
	}, plist.XMLFormat)
	if err != nil {
		t.Fatalf("plist.Marshal: %v", err)
	}
	rows, err := DecodeDefaultsExport(data)
	if err != nil {
		t.Fatalf("DecodeDefaultsExport: %v", err)
	}
	if len(rows) != 4 || rows[0].Key != "Count" || rows[1].Key != "Enabled" {
		t.Fatalf("rows = %#v", rows)
	}
	if rows[1].Type != "bool" || rows[2].Type != "array" || rows[3].Type != "string" {
		t.Fatalf("typed rows = %#v", rows)
	}
}

func TestParseScalarInputValidatesTypes(t *testing.T) {
	tests := []struct {
		kind  string
		input string
		want  string
	}{
		{"bool", "true", "true"},
		{"int", "42", "42"},
		{"float", "1.25", "1.25"},
		{"string", "hello; world", "hello; world"},
	}
	for _, test := range tests {
		flag, value, err := ParseScalarInput(test.kind, test.input)
		if err != nil || flag != "-"+test.kind || value != test.want {
			t.Fatalf("ParseScalarInput(%q, %q) = %q, %q, %v", test.kind, test.input, flag, value, err)
		}
	}
	if _, _, err := ParseScalarInput("int", "1; bad"); err == nil {
		t.Fatal("invalid integer accepted")
	}
}

func TestSummarizePlanCountsRiskAndBlockedItems(t *testing.T) {
	high := tweaks.NewCommandTweak("high", "High", "High", tweaks.CategorySystem, tweaks.RiskHigh, "true", "true", "true", false)
	low := tweaks.NewCommandTweak("low", "Low", "Low", tweaks.CategorySystem, tweaks.RiskLow, "false", "true", "false", false)
	plan := []tweaks.PlanItem{
		{Tweak: high, Action: tweaks.PlanApply},
		{Tweak: low, Action: tweaks.PlanRevert},
		{Tweak: low, Action: tweaks.PlanBlocked},
	}
	summary := SummarizePlan(plan)
	if summary.Apply != 1 || summary.Revert != 1 || summary.Blocked != 1 || summary.HighRisk != 1 {
		t.Fatalf("summary = %#v", summary)
	}
}
