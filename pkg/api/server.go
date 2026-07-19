package api

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

//go:embed web/*
var webFS embed.FS

// StartServer starts the local web server and opens the browser.
func StartServer() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf("127.0.0.1:%s", port)
	token, err := generateCSRFToken()
	if err != nil {
		log.Fatalf("Failed to initialize request security: %v", err)
	}
	handler, err := buildHandler("http://"+addr, token)
	if err != nil {
		log.Fatalf("Failed to initialize web handler: %v", err)
	}

	fmt.Printf("Starting web UI server on http://%s\n", addr)

	// Automatically open the browser
	go func() {
		// Wait just a brief moment to ensure server binds
		exec.Command("open", "http://"+addr).Run()
	}()

	server := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func buildHandler(expectedOrigin, token string) (http.Handler, error) {
	mux := http.NewServeMux()
	danger, err := newDangerSession()
	if err != nil {
		return nil, err
	}
	subFS, err := fs.Sub(webFS, "web")
	if err != nil {
		return nil, err
	}
	mux.Handle("/", http.FileServer(http.FS(subFS)))
	mux.HandleFunc("GET /api/session", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(map[string]string{"csrfToken": token})
	})
	mux.HandleFunc("POST /api/session/dangerous-unlock", danger.handleUnlock)
	mux.HandleFunc("GET /api/tweaks", handleGetTweaks)
	mux.Handle("GET /api/tweaks/dangerous", danger.requireCapability(http.HandlerFunc(handleGetDangerousTweaks)))
	mux.Handle("POST /api/tweaks/{id}/apply", danger.guardHighRiskMutation(http.HandlerFunc(handleApplyTweak)))
	mux.Handle("POST /api/tweaks/{id}/revert", danger.guardHighRiskMutation(http.HandlerFunc(handleRevertTweak)))
	mux.HandleFunc("GET /api/defaults/domains", handleGetDomains)
	mux.HandleFunc("GET /api/defaults/domain/{domain}", handleGetDomainKeys)
	mux.HandleFunc("POST /api/defaults/key", handleWriteKey)
	mux.HandleFunc("DELETE /api/defaults/domain/{domain}/key/{key}", handleDeleteKey)
	mux.HandleFunc("POST /api/defaults/restore", handleRestoreDefaults)
	mux.HandleFunc("GET /api/autoruns", HandleGetAutoruns)
	mux.HandleFunc("POST /api/autoruns/{label}/disable", HandleDisableAutorun)
	mux.HandleFunc("POST /api/autoruns/{label}/enable", HandleEnableAutorun)
	mux.HandleFunc("GET /api/tcpview", HandleGetTCPView)
	mux.HandleFunc("GET /api/processes", HandleGetProcesses)
	mux.HandleFunc("GET /api/processes/{pid}", HandleGetProcessDetail)
	mux.HandleFunc("GET /api/reliability", HandleGetReliability)
	return mutationSecurity(expectedOrigin, token, mux), nil
}
