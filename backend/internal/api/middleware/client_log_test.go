package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMaskToken(t *testing.T) {
	if MaskToken("abcd") != "****" {
		t.Fatalf("short token should be masked")
	}
	masked := MaskToken("1234567890abcdef")
	if masked != "1234...cdef" {
		t.Fatalf("unexpected mask: %s", masked)
	}
}

func TestAPIVersioning(t *testing.T) {
	handler := APIVersioning(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Header().Get("X-API-Version") != APIVersion {
		t.Fatalf("expected default version header")
	}

	req.Header.Set("X-API-Version", "v2")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Header().Get("X-API-Version") != "v2" {
		t.Fatalf("expected echoed version header")
	}
}
