package middleware

import (
	"log"
	"net/http"

	"rionexgate/internal/config"
)

// ClientRegisterAuth gates POST /api/client/register.
// Allowed when:
//   - server.allow_open_register is true (insecure; logs a warning), or
//   - X-API-Key matches server.api_key, or
//   - X-Registration-Secret matches server.registration_secret (when set).
func ClientRegisterAuth(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.Server.AllowOpenRegister {
				log.Printf("warning: allow_open_register=true — unauthenticated device registration is enabled")
				next.ServeHTTP(w, r)
				return
			}
			if SecureEqual(r.Header.Get("X-API-Key"), cfg.Server.APIKey) {
				next.ServeHTTP(w, r)
				return
			}
			secret := cfg.Server.RegistrationSecret
			if secret != "" && SecureEqual(r.Header.Get("X-Registration-Secret"), secret) {
				next.ServeHTTP(w, r)
				return
			}
			http.Error(w, `{"error":"unauthorized","hint":"provide X-API-Key or X-Registration-Secret"}`, http.StatusUnauthorized)
		})
	}
}
