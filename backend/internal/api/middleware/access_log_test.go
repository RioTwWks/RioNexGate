package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSanitizeLogPath(t *testing.T) {
	cases := map[string]string{
		"/api/subscription/abcdef0123456789abcdef0123456789": "/api/subscription/***",
		"/api/subscription/":                                   "/api/subscription/***",
		"/api/users":                                           "/api/users",
		"/api/client/config":                                   "/api/client/config",
		"/":                                                    "/",
	}
	for in, want := range cases {
		if got := SanitizeLogPath(in); got != want {
			t.Errorf("SanitizeLogPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestAccessLoggerMasksSubscriptionToken(t *testing.T) {
	var buf bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(prev)

	secret := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	handler := AccessLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/subscription/"+secret+"?exp=1&sig=deadbeef", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	out := buf.String()
	if strings.Contains(out, secret) {
		t.Fatalf("access log contains raw subscription token: %s", out)
	}
	if strings.Contains(out, "sig=deadbeef") {
		t.Fatalf("access log contains query string: %s", out)
	}
	if !strings.Contains(out, "/api/subscription/***") {
		t.Fatalf("access log missing masked path: %s", out)
	}
}
