package api

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
)

//go:embed web/*
var webFS embed.FS

// StartServer starts the local web server and opens the browser.
func StartServer() {
	mux := http.NewServeMux()

	// Serve the embedded web files
	subFS, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("Failed to create sub filesystem: %v", err)
	}
	mux.Handle("/", http.FileServer(http.FS(subFS)))

	// API Routes (using Go 1.22+ routing features)
	mux.HandleFunc("GET /api/tweaks", handleGetTweaks)
	mux.HandleFunc("POST /api/tweaks/{id}/apply", handleApplyTweak)
	mux.HandleFunc("POST /api/tweaks/{id}/revert", handleRevertTweak)

	mux.HandleFunc("GET /api/defaults/domains", handleGetDomains)
	mux.HandleFunc("GET /api/defaults/domain/{domain}", handleGetDomainKeys)
	mux.HandleFunc("POST /api/defaults/key", handleWriteKey)
	mux.HandleFunc("DELETE /api/defaults/domain/{domain}/key/{key}", handleDeleteKey)
	mux.HandleFunc("POST /api/defaults/restore", handleRestoreDefaults)

	// Sysinternals-style tools
	mux.HandleFunc("GET /api/autoruns", HandleGetAutoruns)
	mux.HandleFunc("POST /api/autoruns/{label}/disable", HandleDisableAutorun)
	mux.HandleFunc("POST /api/autoruns/{label}/enable", HandleEnableAutorun)
	mux.HandleFunc("GET /api/tcpview", HandleGetTCPView)
	mux.HandleFunc("GET /api/processes", HandleGetProcesses)
	mux.HandleFunc("GET /api/processes/{pid}", HandleGetProcessDetail)
	mux.HandleFunc("GET /api/reliability", HandleGetReliability)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf("127.0.0.1:%s", port)

	fmt.Printf("Starting web UI server on http://%s\n", addr)
	
	// Automatically open the browser
	go func() {
		// Wait just a brief moment to ensure server binds
		exec.Command("open", "http://"+addr).Run()
	}()

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
