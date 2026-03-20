package analytics

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"time"
)

// Config holds analytics configuration
type Config struct {
	Enabled      bool
	EndpointURL  string
	UserID       string // Anonymous user identifier
}

// Event represents an analytics event
type Event struct {
	EventType string                 `json:"event_type"`
	Timestamp time.Time              `json:"timestamp"`
	UserID    string                 `json:"user_id"`
	Properties map[string]interface{} `json:"properties"`
	SystemInfo SystemInfo             `json:"system_info"`
}

// SystemInfo contains system metadata
type SystemInfo struct {
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	GoVersion    string `json:"go_version"`
	AppVersion   string `json:"app_version"`
}

var (
	globalConfig Config
	client       = &http.Client{Timeout: 5 * time.Second}
)

// Initialize sets up analytics with configuration
func Initialize(endpointURL string, enabled bool) {
	globalConfig = Config{
		Enabled:     enabled,
		EndpointURL: endpointURL,
		UserID:      getOrCreateUserID(),
	}
}

// Track sends an analytics event
func Track(eventType string, properties map[string]interface{}) {
	if !globalConfig.Enabled || globalConfig.EndpointURL == "" {
		return
	}

	event := Event{
		EventType:  eventType,
		Timestamp:  time.Now(),
		UserID:     globalConfig.UserID,
		Properties: properties,
		SystemInfo: SystemInfo{
			OS:         runtime.GOOS,
			Arch:       runtime.GOARCH,
			GoVersion:  runtime.Version(),
			AppVersion: "1.0.0", // TODO: make this dynamic
		},
	}

	go sendEvent(event)
}

func sendEvent(event Event) {
	payload, err := json.Marshal(event)
	if err != nil {
		return // Silently fail
	}

	req, err := http.NewRequest("POST", globalConfig.EndpointURL, bytes.NewBuffer(payload))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "AfterDarkSystemTools/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return // Silently fail
	}
	defer resp.Body.Close()
}

// getOrCreateUserID generates or retrieves a stable anonymous user ID
func getOrCreateUserID() string {
	// Try to read existing ID from config
	home, err := os.UserHomeDir()
	if err != nil {
		return generateUserID()
	}

	idFile := home + "/.ads-systweak-uid"
	data, err := os.ReadFile(idFile)
	if err == nil && len(data) > 0 {
		return string(data)
	}

	// Generate new ID
	id := generateUserID()
	_ = os.WriteFile(idFile, []byte(id), 0644)
	return id
}

func generateUserID() string {
	// Use hostname + timestamp for anonymous but stable ID
	hostname, _ := os.Hostname()
	return hostname + "-" + time.Now().Format("20060102150405")
}

// Common event types
const (
	EventAppLaunched      = "app_launched"
	EventTweakApplied     = "tweak_applied"
	EventTweakReverted    = "tweak_reverted"
	EventPresetStaged     = "preset_staged"
	EventPlanViewed       = "plan_viewed"
	EventChangesApplied   = "changes_applied"
	EventProTabViewed     = "pro_tab_viewed"
	EventProLinkClicked   = "pro_link_clicked"
	EventSearchUsed       = "search_used"
)

// Helper functions for common events

func TrackAppLaunched(interface_type string) {
	Track(EventAppLaunched, map[string]interface{}{
		"interface": interface_type, // "gui" or "cli"
	})
}

func TrackTweakApplied(tweakID string, category string) {
	Track(EventTweakApplied, map[string]interface{}{
		"tweak_id": tweakID,
		"category": category,
	})
}

func TrackTweakReverted(tweakID string, category string) {
	Track(EventTweakReverted, map[string]interface{}{
		"tweak_id": tweakID,
		"category": category,
	})
}

func TrackPresetStaged(presetName string, tweakCount int) {
	Track(EventPresetStaged, map[string]interface{}{
		"preset_name": presetName,
		"tweak_count": tweakCount,
	})
}

func TrackProLinkClicked(destination string) {
	Track(EventProLinkClicked, map[string]interface{}{
		"destination": destination,
	})
}

func TrackSearchUsed(query string, resultCount int) {
	Track(EventSearchUsed, map[string]interface{}{
		"query_length": len(query),
		"result_count": resultCount,
	})
}
