package api

import (
	"encoding/json"
	"net/http"

	"ads-systweak/pkg/tweaks"
)

type TweakInfo struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Category    tweaks.TweakCategory `json:"category"`
	RiskLevel   tweaks.RiskLevel     `json:"riskLevel"`
	Applied     bool                 `json:"applied"`
	State       tweaks.ProbeState    `json:"state"`
	Error       string               `json:"error,omitempty"`
}

func handleGetTweaks(w http.ResponseWriter, r *http.Request) {
	var list []TweakInfo
	for _, tw := range tweaks.Registry {
		probe := tw.Probe()
		errorText := ""
		if probe.Err != nil {
			errorText = probe.Err.Error()
		}
		list = append(list, TweakInfo{
			ID:          tw.ID(),
			Name:        tw.Name(),
			Description: tw.Description(),
			Category:    tw.Category(),
			RiskLevel:   tw.RiskLevel(),
			Applied:     probe.Applied,
			State:       probe.State,
			Error:       errorText,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func handleApplyTweak(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	for _, tw := range tweaks.Registry {
		if tw.ID() == id {
			probe := tw.Probe()
			if probe.State != tweaks.ProbeApplied && probe.State != tweaks.ProbeOff {
				http.Error(w, "Cannot mutate tweak with state "+string(probe.State)+": "+probeError(probe), http.StatusConflict)
				return
			}
			err := tw.Apply()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	http.Error(w, "Tweak not found", http.StatusNotFound)
}

func handleRevertTweak(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	for _, tw := range tweaks.Registry {
		if tw.ID() == id {
			probe := tw.Probe()
			if probe.State != tweaks.ProbeApplied && probe.State != tweaks.ProbeOff {
				http.Error(w, "Cannot mutate tweak with state "+string(probe.State)+": "+probeError(probe), http.StatusConflict)
				return
			}
			err := tw.Revert()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	http.Error(w, "Tweak not found", http.StatusNotFound)
}

func probeError(probe tweaks.ProbeResult) string {
	if probe.Err == nil {
		return "state could not be determined"
	}
	return probe.Err.Error()
}
