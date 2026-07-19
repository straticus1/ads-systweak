package api

import (
	"strings"
	"testing"
)

func TestDangerZoneWebContract(t *testing.T) {
	htmlBytes, err := webFS.ReadFile("web/index.html")
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	jsBytes, err := webFS.ReadFile("web/app.js")
	if err != nil {
		t.Fatalf("read app script: %v", err)
	}
	html := string(htmlBytes)
	js := string(jsBytes)

	for _, required := range []string{
		`data-tab="danger"`, `id="danger-tab"`, `id="danger-acknowledgement"`,
		`id="danger-phrase"`, `id="danger-unlock"`, `id="danger-tweaks-list"`,
	} {
		if !strings.Contains(html, required) {
			t.Errorf("index.html missing %s", required)
		}
	}
	for _, required := range []string{
		"I KNOW WHAT I AM DOING", "/api/session/dangerous-unlock",
		"/api/tweaks/dangerous", dangerCapabilityHeader,
	} {
		if !strings.Contains(js, required) {
			t.Errorf("app.js missing %q", required)
		}
	}
	for _, forbidden := range []string{"localStorage", "sessionStorage", "document.cookie"} {
		if strings.Contains(js, forbidden) {
			t.Errorf("app.js persists Danger Zone state using %s", forbidden)
		}
	}
}
