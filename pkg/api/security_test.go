package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMutationSecurityRejectsForeignOrigin(t *testing.T) {
	handler := mutationSecurity("http://127.0.0.1:8080", "secret", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:8080/api/tweaks/test/apply", nil)
	req.Header.Set("Origin", "https://evil.example")
	req.Header.Set(csrfHeader, "secret")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, req)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
}

func TestMutationSecurityRejectsMissingToken(t *testing.T) {
	handler := mutationSecurity("http://127.0.0.1:8080", "secret", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodDelete, "http://127.0.0.1:8080/api/defaults/domain/a/key/b", nil)
	req.Header.Set("Origin", "http://127.0.0.1:8080")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, req)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
}

func TestMutationSecurityAllowsSameOriginTokenAndLimitsBody(t *testing.T) {
	var readErr error
	handler := mutationSecurity("http://127.0.0.1:8080", "secret", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, readErr = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:8080/api/defaults/key", strings.NewReader(strings.Repeat("x", maxMutationBody+1)))
	req.Header.Set("Origin", "http://127.0.0.1:8080")
	req.Header.Set(csrfHeader, "secret")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, req)
	if readErr == nil {
		t.Fatal("oversized body read returned nil error")
	}
}

func TestSecurityHeadersAreSet(t *testing.T) {
	handler := mutationSecurity("http://127.0.0.1:8080", "secret", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8080/", nil))

	for _, header := range []string{"Content-Security-Policy", "X-Content-Type-Options", "Referrer-Policy"} {
		if response.Header().Get(header) == "" {
			t.Fatalf("missing security header %s", header)
		}
	}
}
