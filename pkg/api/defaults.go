package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"ads-systweak/pkg/backup"
	"ads-systweak/pkg/tweaks"
	"howett.net/plist"
)

var (
	runCommand         = tweaks.RunCommand
	saveDefaultsBackup = backup.SaveBackup
)

func handleGetDomains(w http.ResponseWriter, r *http.Request) {
	out, err := runCommand("/usr/bin/defaults", "domains")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	domains := strings.Split(out, ",")
	for i := range domains {
		domains[i] = strings.TrimSpace(domains[i])
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(domains)
}

func handleGetDomainKeys(w http.ResponseWriter, r *http.Request) {
	domain := r.PathValue("domain")
	if err := validateDefaultsName(domain); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	out, err := runCommand("/usr/bin/defaults", "export", domain, "-")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var parsed map[string]interface{}
	if err := plist.NewDecoder(bytes.NewReader([]byte(out))).Decode(&parsed); err != nil {
		http.Error(w, "Failed to parse plist: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(parsed)
}

type DefaultsWriteRequest struct {
	Domain string      `json:"domain"`
	Key    string      `json:"key"`
	Type   string      `json:"type"`
	Value  interface{} `json:"value"`
}

func handleWriteKey(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	var req DefaultsWriteRequest
	if err := decoder.Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	if err := validateDefaultsName(req.Domain); err != nil {
		http.Error(w, "invalid domain: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateDefaultsName(req.Key); err != nil {
		http.Error(w, "invalid key: "+err.Error(), http.StatusBadRequest)
		return
	}
	typeFlag, value, err := parseDefaultsValue(req.Type, req.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if tweaks.DryRun {
		w.WriteHeader(http.StatusOK)
		return
	}
	if err := saveDefaultsBackup(req.Domain, req.Key); err != nil {
		http.Error(w, "backup failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := runCommand("/usr/bin/defaults", "write", req.Domain, req.Key, typeFlag, value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func handleDeleteKey(w http.ResponseWriter, r *http.Request) {
	domain := r.PathValue("domain")
	key := r.PathValue("key")
	if err := validateDefaultsName(domain); err != nil {
		http.Error(w, "invalid domain: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateDefaultsName(key); err != nil {
		http.Error(w, "invalid key: "+err.Error(), http.StatusBadRequest)
		return
	}
	if tweaks.DryRun {
		w.WriteHeader(http.StatusOK)
		return
	}
	if err := saveDefaultsBackup(domain, key); err != nil {
		http.Error(w, "backup failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := runCommand("/usr/bin/defaults", "delete", domain, key); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func handleRestoreDefaults(w http.ResponseWriter, r *http.Request) {
	if tweaks.DryRun {
		w.WriteHeader(http.StatusOK)
		return
	}
	if err := backup.RestoreAll(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func validateDefaultsName(value string) error {
	if value == "" {
		return errors.New("value is empty")
	}
	if len(value) > 512 {
		return errors.New("value is too long")
	}
	if strings.ContainsRune(value, 0) {
		return errors.New("value contains NUL")
	}
	return nil
}

func parseDefaultsValue(valueType string, raw interface{}) (string, string, error) {
	switch strings.ToLower(valueType) {
	case "bool", "boolean":
		value, ok := raw.(bool)
		if !ok {
			return "", "", errors.New("bool value must be true or false")
		}
		return "-bool", strconv.FormatBool(value), nil
	case "int", "integer":
		number, ok := raw.(json.Number)
		if !ok {
			return "", "", errors.New("int value must be a JSON integer")
		}
		value, err := strconv.ParseInt(number.String(), 10, 64)
		if err != nil {
			return "", "", errors.New("int value is invalid or out of range")
		}
		return "-int", strconv.FormatInt(value, 10), nil
	case "float":
		number, ok := raw.(json.Number)
		if !ok {
			return "", "", errors.New("float value must be a JSON number")
		}
		value, err := strconv.ParseFloat(number.String(), 64)
		if err != nil || math.IsInf(value, 0) || math.IsNaN(value) {
			return "", "", errors.New("float value is invalid or out of range")
		}
		return "-float", strconv.FormatFloat(value, 'g', -1, 64), nil
	case "string":
		value, ok := raw.(string)
		if !ok {
			return "", "", errors.New("string value must be a JSON string")
		}
		if strings.ContainsRune(value, 0) {
			return "", "", errors.New("string value contains NUL")
		}
		return "-string", value, nil
	default:
		return "", "", fmt.Errorf("unsupported defaults type %q", valueType)
	}
}
