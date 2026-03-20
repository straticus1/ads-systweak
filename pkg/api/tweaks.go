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
}

func handleGetTweaks(w http.ResponseWriter, r *http.Request) {
	var list []TweakInfo
	for _, tw := range tweaks.Registry {
		actual, err := tw.IsApplied()
		if err != nil {
			actual = false
		}
		list = append(list, TweakInfo{
			ID:          tw.ID(),
			Name:        tw.Name(),
			Description: tw.Description(),
			Category:    tw.Category(),
			RiskLevel:   tw.RiskLevel(),
			Applied:     actual,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func handleApplyTweak(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	for _, tw := range tweaks.Registry {
		if tw.ID() == id {
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
