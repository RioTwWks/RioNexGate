package middleware

import (
	"log"
	"net/http"
	"strconv"
)

func MaskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "..." + token[len(token)-4:]
}

func ClientRequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Device-Token")
		userID := ""
		if dc, ok := DeviceFromContext(r.Context()); ok && dc.User != nil {
			userID = strconv.FormatUint(uint64(dc.User.ID), 10)
		}
		log.Printf("client_api method=%s path=%s device_token=%s user_id=%s",
			r.Method, r.URL.Path, MaskToken(token), userID)
		next.ServeHTTP(w, r)
	})
}
