package middleware

import (
	"context"
	"log"
	"net/http"

	"rionexgate/internal/config"
	"rionexgate/internal/db"
	"rionexgate/internal/models"
)

type registerAuthKey struct{}

type RegisterAuth struct {
	Mode   string // "api_key" | "registration_secret" | "open" | "invite"
	Invite *models.Invite
}

func RegisterAuthFromContext(ctx context.Context) (*RegisterAuth, bool) {
	v, ok := ctx.Value(registerAuthKey{}).(*RegisterAuth)
	return v, ok
}

// ResolveRegisterAuth decides whether a client registration request is allowed.
// inviteToken may come from X-Invite-Token or the JSON body field invite_token.
func ResolveRegisterAuth(cfg *config.Config, database *db.DB, apiKey, regSecret, inviteToken string) *RegisterAuth {
	if cfg != nil && cfg.Server.AllowOpenRegister {
		log.Printf("warning: allow_open_register=true — unauthenticated device registration is enabled")
		return &RegisterAuth{Mode: "open"}
	}
	if cfg != nil && SecureEqual(apiKey, cfg.APIKey()) {
		return &RegisterAuth{Mode: "api_key"}
	}
	if cfg != nil {
		secret := cfg.Server.RegistrationSecret
		if secret != "" && SecureEqual(regSecret, secret) {
			return &RegisterAuth{Mode: "registration_secret"}
		}
	}
	if inviteToken != "" && database != nil {
		inv, err := database.GetInviteByToken(inviteToken)
		if err == nil && inv.Usable() {
			return &RegisterAuth{Mode: "invite", Invite: inv}
		}
	}
	return nil
}

// ClientRegisterAuth gates POST /api/client/register using headers only.
// Invite tokens in the JSON body are validated inside the register handler.
func ClientRegisterAuth(cfg *config.Config, database *db.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := ResolveRegisterAuth(cfg, database,
				r.Header.Get("X-API-Key"),
				r.Header.Get("X-Registration-Secret"),
				r.Header.Get("X-Invite-Token"),
			)
			if auth == nil {
				// Allow through so the handler can accept invite_token from JSON body.
				next.ServeHTTP(w, r)
				return
			}
			ctx := context.WithValue(r.Context(), registerAuthKey{}, auth)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
