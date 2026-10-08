package middleware

import (
	"crypto/subtle"
	"net/http"

	"rionexgate/internal/config"
)

func APIKeyAuth(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			provided := r.Header.Get("X-API-Key")
			if !SecureEqual(provided, apiKey) {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// APIKeyAuthDynamic reads the API key from cfg on each request (supports rotation).
func APIKeyAuthDynamic(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			provided := r.Header.Get("X-API-Key")
			expected := ""
			if cfg != nil {
				expected = cfg.APIKey()
			}
			if !SecureEqual(provided, expected) {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// SecureEqual compares secrets in constant time. Empty expected never matches.
func SecureEqual(provided, expected string) bool {
	if expected == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}
