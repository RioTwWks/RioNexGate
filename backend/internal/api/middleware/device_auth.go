package middleware

import (
	"context"
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

func DeviceTokenAuth(database *db.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("X-Device-Token")
			if token == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			device, err := database.GetDeviceByToken(token)
			if err != nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			user, err := database.GetUser(device.UserID)
			if err != nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			if !user.Active {
				http.Error(w, `{"error":"user inactive"}`, http.StatusForbidden)
				return
			}
			dc := &DeviceContext{Device: device, User: user}
			ctx := context.WithValue(r.Context(), deviceContextKey{}, dc)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
