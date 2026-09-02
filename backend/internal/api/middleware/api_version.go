package middleware

import (
	"net/http"
)

const APIVersion = "v1"

func APIVersioning(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		version := r.Header.Get("X-API-Version")
		if version == "" {
			version = APIVersion
		}
		w.Header().Set("X-API-Version", version)
		next.ServeHTTP(w, r)
	})
}
