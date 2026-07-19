package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ads-systweak/pkg/tweaks"
)

type dangerTestTweak struct {
	id      string
	risk    tweaks.RiskLevel
	applied bool
	applies int
	reverts int
}

func (t *dangerTestTweak) ID() string                     { return t.id }
func (t *dangerTestTweak) Name() string                   { return t.id }
func (t *dangerTestTweak) Description() string            { return t.id + " description" }
func (t *dangerTestTweak) Category() tweaks.TweakCategory { return tweaks.CategoryOther }
func (t *dangerTestTweak) RiskLevel() tweaks.RiskLevel    { return t.risk }
func (t *dangerTestTweak) Probe() tweaks.ProbeResult {
	state := tweaks.ProbeOff
	if t.applied {
		state = tweaks.ProbeApplied
	}
	return tweaks.ProbeResult{State: state, Applied: t.applied}
}
func (t *dangerTestTweak) IsApplied() (bool, error) { return t.applied, nil }
func (t *dangerTestTweak) Apply() error {
	t.applies++
	t.applied = true
	return nil
}
func (t *dangerTestTweak) Revert() error {
	t.reverts++
	t.applied = false
	return nil
}

func dangerTestHandler(t *testing.T, registry []tweaks.Tweak) http.Handler {
	t.Helper()
	previous := tweaks.Registry
	tweaks.Registry = registry
	t.Cleanup(func() { tweaks.Registry = previous })
	handler, err := buildHandler("http://127.0.0.1:8080", "csrf-secret")
	if err != nil {
		t.Fatalf("buildHandler: %v", err)
	}
	return handler
}

func performDangerRequest(handler http.Handler, method, target string, body []byte, dangerCapability string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "http://127.0.0.1:8080"+target, bytes.NewReader(body))
	if method != http.MethodGet {
		req.Header.Set("Origin", "http://127.0.0.1:8080")
		req.Header.Set(csrfHeader, "csrf-secret")
		req.Header.Set("Content-Type", "application/json")
	}
	if dangerCapability != "" {
		req.Header.Set(dangerCapabilityHeader, dangerCapability)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	return response
}

func unlockDangerCapability(t *testing.T, handler http.Handler) string {
	t.Helper()
	body, _ := json.Marshal(map[string]interface{}{
		"acknowledged": true,
		"phrase":       tweaks.DangerUnlockPhrase,
	})
	response := performDangerRequest(handler, http.MethodPost, "/api/session/dangerous-unlock", body, "")
	if response.Code != http.StatusOK {
		t.Fatalf("unlock status = %d; body = %s", response.Code, response.Body.String())
	}
	var payload struct {
		Capability string `json:"capability"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode unlock: %v", err)
	}
	if payload.Capability == "" {
		t.Fatal("unlock returned empty capability")
	}
	return payload.Capability
}

func TestTweaksEndpointHidesHighRisk(t *testing.T) {
	low := &dangerTestTweak{id: "low", risk: tweaks.RiskLow}
	high := &dangerTestTweak{id: "secret-high", risk: tweaks.RiskHigh}
	handler := dangerTestHandler(t, []tweaks.Tweak{low, high})

	response := performDangerRequest(handler, http.MethodGet, "/api/tweaks", nil, "")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	if bytes.Contains(response.Body.Bytes(), []byte("secret-high")) {
		t.Fatalf("ordinary response exposed High-risk tweak: %s", response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("low")) {
		t.Fatalf("ordinary response omitted Low-risk tweak: %s", response.Body.String())
	}
}

func TestDangerousTweaksEndpointRequiresUnlockedCapability(t *testing.T) {
	high := &dangerTestTweak{id: "high", risk: tweaks.RiskHigh}
	handler := dangerTestHandler(t, []tweaks.Tweak{high})

	locked := performDangerRequest(handler, http.MethodGet, "/api/tweaks/dangerous", nil, "")
	if locked.Code != http.StatusForbidden {
		t.Fatalf("locked status = %d, want 403", locked.Code)
	}

	capability := unlockDangerCapability(t, handler)
	unlocked := performDangerRequest(handler, http.MethodGet, "/api/tweaks/dangerous", nil, capability)
	if unlocked.Code != http.StatusOK || !bytes.Contains(unlocked.Body.Bytes(), []byte("high")) {
		t.Fatalf("unlocked response = %d %s", unlocked.Code, unlocked.Body.String())
	}
}

func TestDangerUnlockRejectsIncompleteAcknowledgement(t *testing.T) {
	handler := dangerTestHandler(t, nil)
	for _, body := range [][]byte{
		[]byte(`{"acknowledged":false,"phrase":"I KNOW WHAT I AM DOING"}`),
		[]byte(`{"acknowledged":true,"phrase":"I know what I am doing"}`),
	} {
		response := performDangerRequest(handler, http.MethodPost, "/api/session/dangerous-unlock", body, "")
		if response.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403; body = %s", response.Code, response.Body.String())
		}
	}
}

func TestDangerousMutationsRequireCapabilityButOrdinaryMutationsDoNot(t *testing.T) {
	low := &dangerTestTweak{id: "low", risk: tweaks.RiskLow}
	high := &dangerTestTweak{id: "high", risk: tweaks.RiskHigh}
	handler := dangerTestHandler(t, []tweaks.Tweak{low, high})

	locked := performDangerRequest(handler, http.MethodPost, "/api/tweaks/high/apply", nil, "")
	if locked.Code != http.StatusForbidden || high.applies != 0 {
		t.Fatalf("locked dangerous mutation = %d, apply calls = %d", locked.Code, high.applies)
	}

	ordinary := performDangerRequest(handler, http.MethodPost, "/api/tweaks/low/apply", nil, "")
	if ordinary.Code != http.StatusOK || low.applies != 1 {
		t.Fatalf("ordinary mutation = %d, apply calls = %d", ordinary.Code, low.applies)
	}

	capability := unlockDangerCapability(t, handler)
	apply := performDangerRequest(handler, http.MethodPost, "/api/tweaks/high/apply", nil, capability)
	revert := performDangerRequest(handler, http.MethodPost, "/api/tweaks/high/revert", nil, capability)
	if apply.Code != http.StatusOK || revert.Code != http.StatusOK || high.applies != 1 || high.reverts != 1 {
		t.Fatalf("unlocked apply/revert = %d/%d, calls = %d/%d", apply.Code, revert.Code, high.applies, high.reverts)
	}
}
