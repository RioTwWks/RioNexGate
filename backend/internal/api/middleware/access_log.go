package middleware

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// AccessLogger logs method, sanitized path, status, and duration.
// Query strings are omitted and credential-bearing path segments
// (e.g. /api/subscription/{token}) are masked, so tokens never hit access logs.
func AccessLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(ww, r)
		log.Printf("%s %s %d %s", r.Method, SanitizeLogPath(r.URL.Path), ww.status, time.Since(start))
	})
}

// SanitizeLogPath masks credential-bearing path segments before logging.
// The subscription token is a permanent bearer credential that lives in the
// URL path; logging it in clear text would hand full account access to anyone
// reading the logs.
func SanitizeLogPath(path string) string {
	const subscriptionPrefix = "/api/subscription/"
	if strings.HasPrefix(path, subscriptionPrefix) {
		return subscriptionPrefix + "***"
	}
	return path
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *statusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

// Ensure optional interfaces are preserved for SSE / hijack.
var (
	_ http.Flusher  = (*statusWriter)(nil)
	_ http.Hijacker = (*statusWriter)(nil)
)
