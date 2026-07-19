package api

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"ads-systweak/pkg/tweaks"
)

const dangerCapabilityHeader = "X-ADS-Danger-Unlock"

type dangerSession struct {
	capability string
}

func newDangerSession() (*dangerSession, error) {
	capability, err := generateCSRFToken()
	if err != nil {
		return nil, err
	}
	return &dangerSession{capability: capability}, nil
}

func (s *dangerSession) authorized(r *http.Request) bool {
	provided := r.Header.Get(dangerCapabilityHeader)
	return len(provided) == len(s.capability) && subtle.ConstantTimeCompare([]byte(provided), []byte(s.capability)) == 1
}

func (s *dangerSession) handleUnlock(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Acknowledged bool   `json:"acknowledged"`
		Phrase       string `json:"phrase"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "invalid unlock request", http.StatusBadRequest)
		return
	}
	if err := ensureJSONEnd(decoder); err != nil {
		http.Error(w, "invalid unlock request", http.StatusBadRequest)
		return
	}
	if !tweaks.ValidateDangerUnlock(request.Acknowledged, request.Phrase) {
		http.Error(w, "Danger Zone acknowledgement did not match", http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(map[string]string{"capability": s.capability})
}

func ensureJSONEnd(decoder *json.Decoder) error {
	var extra interface{}
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return errors.New("multiple JSON values")
	}
	return err
}

func (s *dangerSession) requireCapability(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.authorized(r) {
			http.Error(w, "Danger Zone is locked", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *dangerSession) guardHighRiskMutation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		for _, tweak := range tweaks.Registry {
			if tweak.ID() == id && tweak.RiskLevel() == tweaks.RiskHigh && !s.authorized(r) {
				http.Error(w, "Danger Zone is locked", http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
