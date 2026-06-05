package api

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ReliabilityEvent represents a single crash/hang/panic report.
type ReliabilityEvent struct {
	Timestamp string `json:"timestamp"` // ISO8601 / RFC3339
	Type      string `json:"type"`      // "Crash", "Hang", "Panic", "Spin", "Wakeup"
	AppName   string `json:"appName"`
	Version   string `json:"version"`
	FilePath  string `json:"filePath"`
	Summary   string `json:"summary"`
	OS        string `json:"os"`
}

// HandleGetReliability scans macOS diagnostic report directories and returns
// a JSON array of ReliabilityEvent, sorted newest-first, capped at 200.
func HandleGetReliability(w http.ResponseWriter, r *http.Request) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		http.Error(w, "cannot determine home directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	dirs := []string{
		filepath.Join(homeDir, "Library", "Logs", "DiagnosticReports"),
		"/Library/Logs/DiagnosticReports",
	}

	var events []ReliabilityEvent

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			// Directory missing or inaccessible — skip silently.
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			ext := strings.ToLower(filepath.Ext(name))

			eventType := extensionToType(ext)
			if eventType == "" {
				continue
			}

			appName, ts := parseFilename(name, ext)
			if appName == "" {
				appName = strings.TrimSuffix(name, filepath.Ext(name))
			}

			fullPath := filepath.Join(dir, name)
			summary, osVer := readReportHead(fullPath, eventType)

			events = append(events, ReliabilityEvent{
				Timestamp: ts,
				Type:      eventType,
				AppName:   appName,
				FilePath:  fullPath,
				Summary:   summary,
				OS:        osVer,
			})
		}
	}

	// Sort newest-first.
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp > events[j].Timestamp
	})

	if len(events) > 200 {
		events = events[:200]
	}

	// Guarantee a non-null JSON array.
	if events == nil {
		events = []ReliabilityEvent{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// extensionToType maps a file extension to a human-readable event type.
// Returns "" for unrecognised extensions.
func extensionToType(ext string) string {
	switch ext {
	case ".crash", ".ips":
		return "Crash"
	case ".hang":
		return "Hang"
	case ".panic":
		return "Panic"
	case ".spindump":
		return "Spin"
	case ".wakeup":
		return "Wakeup"
	default:
		return ""
	}
}

// parseFilename extracts the app name and an RFC3339 timestamp string from a
// macOS diagnostic report filename.
//
// Expected pattern: AppName_YYYY-MM-DD-HHMMSS_hostname.ext
// or (wakeup):     AppName_YYYY-MM-DD-HHMMSS.ext
func parseFilename(name, ext string) (appName, timestamp string) {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	parts := strings.Split(base, "_")

	// Walk parts to find the date segment (starts with a 4-digit year).
	for i, p := range parts {
		if len(p) >= 4 && isDigit4(p[:4]) && strings.Count(p, "-") >= 2 {
			// Everything before index i is the app name.
			appName = strings.Join(parts[:i], "_")

			t, err := time.Parse("2006-01-02-150405", p)
			if err == nil {
				timestamp = t.Format(time.RFC3339)
			}
			return
		}
	}

	// Fallback: use the whole base as appName, empty timestamp.
	appName = base
	return
}

// isDigit4 returns true when all four bytes of s are ASCII digits.
func isDigit4(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// readReportHead opens the file and reads up to 50 lines, extracting a
// one-line summary and the macOS version string.
func readReportHead(path, eventType string) (summary, osVer string) {
	f, err := os.Open(path)
	if err != nil {
		return "(permission denied)", ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineCount := 0
	var firstContent string

	for scanner.Scan() && lineCount < 50 {
		line := scanner.Text()
		lineCount++

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Track the first non-empty line as a fallback summary.
		if firstContent == "" {
			firstContent = trimmed
		}

		// Extract OS version from any report type.
		if osVer == "" && strings.HasPrefix(trimmed, "OS Version:") {
			osVer = strings.TrimSpace(strings.TrimPrefix(trimmed, "OS Version:"))
		}

		// Extract type-specific summary lines (prefer first match).
		if summary == "" {
			switch eventType {
			case "Crash":
				if strings.HasPrefix(trimmed, "Exception Type:") || strings.HasPrefix(trimmed, "Crashed Thread:") {
					summary = trimmed
				}
			case "Hang":
				if strings.HasPrefix(trimmed, "Duration:") {
					summary = trimmed
				}
			case "Panic":
				if strings.HasPrefix(trimmed, "panic(cpu") || strings.HasPrefix(trimmed, "Kernel trap") {
					summary = trimmed
				}
			}
		}

		// Stop early once we have everything we need.
		if summary != "" && osVer != "" {
			break
		}
	}

	if summary == "" {
		summary = firstContent
	}
	return
}
