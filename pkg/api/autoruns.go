package api

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"howett.net/plist"
)

// AutorunItem represents a single startup item of any type.
type AutorunItem struct {
	Label     string `json:"label"`
	Type      string `json:"type"`
	Path      string `json:"path"`
	Program   string `json:"program"`
	Enabled   bool   `json:"enabled"`
	RunAtLoad bool   `json:"runAtLoad"`
	Source    string `json:"source"`
	BundleID  string `json:"bundleId"`
	ReadOnly  bool   `json:"readOnly"`
}

type launchPlist struct {
	Label            string   `plist:"Label"`
	Program          string   `plist:"Program"`
	ProgramArguments []string `plist:"ProgramArguments"`
	RunAtLoad        bool     `plist:"RunAtLoad"`
}

var (
	autorunCacheMu sync.RWMutex
	autorunCache   map[string]AutorunItem
)

// HandleGetAutoruns aggregates all startup items and returns them as JSON.
func HandleGetAutoruns(w http.ResponseWriter, r *http.Request) {
	items := collectAutoruns()

	newCache := make(map[string]AutorunItem, len(items))
	for _, item := range items {
		newCache[item.Label] = item
	}
	autorunCacheMu.Lock()
	autorunCache = newCache
	autorunCacheMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// HandleDisableAutorun disables the autorun item identified by the URL label.
func HandleDisableAutorun(w http.ResponseWriter, r *http.Request) {
	label := r.PathValue("label")
	if label == "" {
		http.Error(w, "missing label", http.StatusBadRequest)
		return
	}

	autorunCacheMu.RLock()
	item, ok := autorunCache[label]
	autorunCacheMu.RUnlock()

	if !ok {
		http.Error(w, "autorun item not found", http.StatusNotFound)
		return
	}

	if item.ReadOnly {
		http.Error(w, "system item is read-only", http.StatusForbidden)
		return
	}

	switch item.Type {
	case "LaunchAgent", "LaunchDaemon":
		if item.Path == "" || item.Path == "N/A" {
			http.Error(w, "no plist path available", http.StatusBadRequest)
			return
		}
		out, err := exec.Command("launchctl", "unload", "-w", item.Path).CombinedOutput()
		if err != nil {
			http.Error(w, "launchctl unload failed: "+string(out), http.StatusInternalServerError)
			return
		}
	case "LoginItem":
		script := `tell application "System Events" to delete login item "` + item.Label + `"`
		out, err := exec.Command("osascript", "-e", script).CombinedOutput()
		if err != nil {
			http.Error(w, "osascript failed: "+string(out), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "disable not supported for type "+item.Type, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleEnableAutorun enables the autorun item identified by the URL label.
func HandleEnableAutorun(w http.ResponseWriter, r *http.Request) {
	label := r.PathValue("label")
	if label == "" {
		http.Error(w, "missing label", http.StatusBadRequest)
		return
	}

	autorunCacheMu.RLock()
	item, ok := autorunCache[label]
	autorunCacheMu.RUnlock()

	if !ok {
		http.Error(w, "autorun item not found", http.StatusNotFound)
		return
	}

	if item.ReadOnly {
		http.Error(w, "system item is read-only", http.StatusForbidden)
		return
	}

	switch item.Type {
	case "LaunchAgent", "LaunchDaemon":
		if item.Path == "" || item.Path == "N/A" {
			http.Error(w, "no plist path available", http.StatusBadRequest)
			return
		}
		out, err := exec.Command("launchctl", "load", "-w", item.Path).CombinedOutput()
		if err != nil {
			http.Error(w, "launchctl load failed: "+string(out), http.StatusInternalServerError)
			return
		}
	case "LoginItem":
		http.Error(w, "re-adding login items via API is not supported (requires app path)", http.StatusBadRequest)
		return
	default:
		http.Error(w, "enable not supported for type "+item.Type, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// collectAutoruns gathers all startup items from all sources.
func collectAutoruns() []AutorunItem {
	var items []AutorunItem
	items = append(items, collectLaunchItems()...)
	items = append(items, collectLoginItems()...)
	items = append(items, collectCronJobs()...)
	items = append(items, collectKernelExtensions()...)
	return items
}

func collectLaunchItems() []AutorunItem {
	home, _ := os.UserHomeDir()

	dirs := []struct {
		path     string
		itemType string
		readOnly bool
	}{
		{filepath.Join(home, "Library", "LaunchAgents"), "LaunchAgent", false},
		{"/Library/LaunchAgents", "LaunchAgent", false},
		{"/Library/LaunchDaemons", "LaunchDaemon", false},
		{"/System/Library/LaunchDaemons", "LaunchDaemon", true},
		{"/System/Library/LaunchAgents", "LaunchAgent", true},
	}

	var items []AutorunItem

	for _, d := range dirs {
		entries, err := os.ReadDir(d.path)
		if err != nil {
			// Directory may not exist; skip silently.
			continue
		}

		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".plist") {
				continue
			}

			plistPath := filepath.Join(d.path, e.Name())
			data, err := os.ReadFile(plistPath)
			if err != nil {
				continue
			}

			var lp launchPlist
			_, err = plist.Unmarshal(data, &lp)
			if err != nil {
				// Fall back to using the filename as the label.
				lp.Label = strings.TrimSuffix(e.Name(), ".plist")
			}

			label := lp.Label
			if label == "" {
				label = strings.TrimSuffix(e.Name(), ".plist")
			}

			program := lp.Program
			if program == "" && len(lp.ProgramArguments) > 0 {
				program = lp.ProgramArguments[0]
			}

			enabled := isLaunchctlLoaded(label)

			items = append(items, AutorunItem{
				Label:     label,
				Type:      d.itemType,
				Path:      plistPath,
				Program:   program,
				Enabled:   enabled,
				RunAtLoad: lp.RunAtLoad,
				Source:    d.path,
				ReadOnly:  d.readOnly,
			})
		}
	}

	return items
}

// isLaunchctlLoaded checks whether a service label is currently loaded.
func isLaunchctlLoaded(label string) bool {
	err := exec.Command("launchctl", "list", label).Run()
	return err == nil
}

func collectLoginItems() []AutorunItem {
	script := `tell application "System Events" to get the name of every login item`
	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return nil
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil
	}

	var items []AutorunItem
	for _, name := range strings.Split(raw, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		items = append(items, AutorunItem{
			Label:   name,
			Type:    "LoginItem",
			Path:    "N/A",
			Program: name,
			Enabled: true,
			Source:  "Login Items",
		})
	}
	return items
}

func collectCronJobs() []AutorunItem {
	out, err := exec.Command("crontab", "-l").Output()
	if err != nil {
		// No crontab or crontab -l returned non-zero (no crontab for user).
		return nil
	}

	var items []AutorunItem
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// A cron line has 5 time fields followed by the command.
		fields := strings.Fields(line)
		command := ""
		if len(fields) > 5 {
			command = strings.Join(fields[5:], " ")
		} else if len(fields) > 0 {
			command = fields[len(fields)-1]
		}

		items = append(items, AutorunItem{
			Label:   line,
			Type:    "CronJob",
			Path:    "N/A",
			Program: command,
			Enabled: true,
			Source:  "crontab",
		})
	}
	return items
}

func collectKernelExtensions() []AutorunItem {
	var items []AutorunItem
	if out, _ := exec.Command("/usr/bin/kmutil", "showloaded").CombinedOutput(); len(out) > 0 {
		items = append(items, parseKMUtilOutput(string(out))...)
	}
	if out, _ := exec.Command("/usr/bin/systemextensionsctl", "list").CombinedOutput(); len(out) > 0 {
		items = append(items, parseSystemExtensionsOutput(string(out))...)
	}
	return items
}

func parseKMUtilOutput(output string) []AutorunItem {
	var items []AutorunItem
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 2 {
			continue
		}
		name := ""
		for _, field := range fields {
			if strings.Contains(field, ".") && !strings.HasPrefix(field, "0x") && !strings.Contains(field, "=") {
				name = strings.TrimSpace(strings.SplitN(field, "(", 2)[0])
				break
			}
		}
		if name == "" || name == "Name" {
			continue
		}
		items = append(items, AutorunItem{Label: name, Type: "KernelExtension", Path: "N/A", Program: name, Enabled: true, Source: "kmutil", ReadOnly: true})
	}
	return items
}

func parseSystemExtensionsOutput(output string) []AutorunItem {
	var items []AutorunItem
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.Contains(trimmed, "[") || strings.HasPrefix(trimmed, "---") {
			continue
		}
		fields := strings.Fields(trimmed)
		bundleID := ""
		for _, field := range fields {
			if strings.Contains(field, ".") && !strings.HasPrefix(field, "[") {
				bundleID = field
				break
			}
		}
		if bundleID == "" || bundleID == "bundleID" {
			continue
		}
		enabled := strings.HasPrefix(trimmed, "* *") && strings.Contains(strings.ToLower(trimmed), "activated enabled")
		items = append(items, AutorunItem{Label: bundleID, Type: "SystemExtension", Path: "N/A", Program: bundleID, Enabled: enabled, Source: "systemextensionsctl", ReadOnly: true})
	}
	return items
}
