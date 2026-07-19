package api

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"
)

const (
	csrfHeader      = "X-ADS-CSRF"
	maxMutationBody = 64 << 10
)

func generateCSRFToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

func mutationSecurity(expectedOrigin, token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; object-src 'none'; base-uri 'none'; frame-ancestors 'none'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Frame-Options", "DENY")
		if strings.HasPrefix(r.URL.Path, "/api/") && isMutation(r.Method) {
			origin := r.Header.Get("Origin")
			if origin != "" && origin != expectedOrigin {
				http.Error(w, "foreign origin rejected", http.StatusForbidden)
				return
			}
			provided := r.Header.Get(csrfHeader)
			if len(provided) != len(token) || subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
				http.Error(w, "missing or invalid CSRF token", http.StatusForbidden)
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, maxMutationBody)
		}
		next.ServeHTTP(w, r)
	})
}

func isMutation(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}
