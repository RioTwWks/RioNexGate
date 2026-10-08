package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"rionexgate/internal/db"
	"rionexgate/internal/models"
)

type deviceContextKey struct{}

type DeviceContext struct {
	Device *models.Device
	User   *models.User
}

func DeviceFromContext(ctx context.Context) (*DeviceContext, bool) {
	dc, ok := ctx.Value(deviceContextKey{}).(*DeviceContext)
	return dc, ok
}

func deviceTokenFromRequest(r *http.Request) string {
	return r.Header.Get("X-Device-Token")
}

func writeDeviceAuthError(w http.ResponseWriter, status int, msg, hint string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	payload := map[string]string{"error": msg}
	if hint != "" {
		payload["hint"] = hint
	}
	_ = json.NewEncoder(w).Encode(payload)
}

func userAccessForbidden(user *models.User) (forbidden bool, msg string) {
	if user == nil {
		return true, "unauthorized"
	}
	if !user.Active {
		return true, "user inactive"
	}
	if user.IsExpired() {
		return true, "user expired"
	}
	if user.IsQuotaExceeded() {
		return true, "traffic quota exceeded"
	}
	return false, ""
}

func DeviceTokenAuth(database *db.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := deviceTokenFromRequest(r)
			if token == "" {
				writeDeviceAuthError(w, http.StatusUnauthorized, "unauthorized",
					"RioNexTunnel requires X-Device-Token header (from POST /api/client/register), not subscription token")
				return
			}
			device, err := database.GetDeviceByToken(token)
			if err != nil {
				if user, subErr := database.GetUserBySubscriptionToken(token); subErr == nil && user != nil {
					writeDeviceAuthError(w, http.StatusUnauthorized, "subscription token cannot be used for client config",
						"Use device_token from POST /api/client/register or GET /api/users/{id}/devices (X-API-Key)")
					return
				}
				writeDeviceAuthError(w, http.StatusUnauthorized, "invalid device token",
					"Register the device via POST /api/client/register or list tokens at GET /api/users/{id}/devices")
				return
			}
			user, err := database.GetUser(device.UserID)
			if err != nil {
				writeDeviceAuthError(w, http.StatusUnauthorized, "unauthorized", "")
				return
			}
			if forbidden, msg := userAccessForbidden(user); forbidden {
				writeDeviceAuthError(w, http.StatusForbidden, msg, "")
				return
			}
			dc := &DeviceContext{Device: device, User: user}
			ctx := context.WithValue(r.Context(), deviceContextKey{}, dc)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
