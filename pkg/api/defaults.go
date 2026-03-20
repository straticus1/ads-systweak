package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"ads-systweak/pkg/backup"
	"ads-systweak/pkg/tweaks"
	"howett.net/plist"
)

func handleGetDomains(w http.ResponseWriter, r *http.Request) {
	out, err := tweaks.RunShell("defaults domains")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	domains := strings.Split(out, ",")
	for i := range domains {
		domains[i] = strings.TrimSpace(domains[i])
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domains)
}

func handleGetDomainKeys(w http.ResponseWriter, r *http.Request) {
	domain := r.PathValue("domain")

	// Check if string is safe (prevent injection)
	if strings.ContainsAny(domain, " \t\n\r$;|&><") {
		http.Error(w, "Invalid domain name", http.StatusBadRequest)
		return
	}

	out, err := tweaks.RunShell(fmt.Sprintf("defaults export %s -", domain))
	if err != nil {
		// some domains may just be empty or fail to read
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var parsed map[string]interface{}
	decoder := plist.NewDecoder(bytes.NewReader([]byte(out)))
	err = decoder.Decode(&parsed)
	if err != nil {
		http.Error(w, "Failed to parse plist: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(parsed)
}

type DefaultsWriteRequest struct {
	Domain string      `json:"domain"`
	Key    string      `json:"key"`
	Type   string      `json:"type"` // string, int, float, bool, dict, array
	Value  interface{} `json:"value"`
}

func handleWriteKey(w http.ResponseWriter, r *http.Request) {
	var req DefaultsWriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if strings.ContainsAny(req.Domain, " \t\n\r$;|&><") || strings.ContainsAny(req.Key, " \t\n\r$;|&><") {
		http.Error(w, "Invalid input characters", http.StatusBadRequest)
		return
	}

	valStr := fmt.Sprintf("%v", req.Value)
	switch req.Type {
	case "bool", "boolean":
		if b, ok := req.Value.(bool); ok {
			if b {
				valStr = "true"
			} else {
				valStr = "false"
			}
		}
		req.Type = "-bool"
	case "int", "integer":
		req.Type = "-int"
	case "float":
		req.Type = "-float"
	case "string":
		req.Type = "-string"
		valStr = fmt.Sprintf("'%s'", strings.ReplaceAll(valStr, "'", "'\\''"))
	default:
		// Unsupported complex writes for now via simple CLI
		http.Error(w, "Unsupported type for writing via API", http.StatusBadRequest)
		return
	}

	if !tweaks.DryRun {
		backup.SaveBackup(req.Domain, req.Key)
	}

	cmd := fmt.Sprintf("defaults write %s %s %s %s", req.Domain, req.Key, req.Type, valStr)
	_, err := tweaks.RunShell(cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleDeleteKey(w http.ResponseWriter, r *http.Request) {
	domain := r.PathValue("domain")
	key := r.PathValue("key")

	if strings.ContainsAny(domain, " \t\n\r$;|&><") || strings.ContainsAny(key, " \t\n\r$;|&><") {
		http.Error(w, "Invalid input characters", http.StatusBadRequest)
		return
	}

	if !tweaks.DryRun {
		backup.SaveBackup(domain, key)
	}

	cmd := fmt.Sprintf("defaults delete %s %s", domain, key)
	_, err := tweaks.RunShell(cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleRestoreDefaults(w http.ResponseWriter, r *http.Request) {
	if tweaks.DryRun {
		fmt.Println("[DRY RUN] Would restore all defaults from backups")
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := backup.RestoreAll(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
