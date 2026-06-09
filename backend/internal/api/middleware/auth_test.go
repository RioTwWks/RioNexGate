package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIKeyAuth(t *testing.T) {
	called := false
	handler := APIKeyAuth("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "wrong")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}

	req.Header.Set("X-API-Key", "secret")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if !called || rr.Code != http.StatusOK {
		t.Fatalf("expected authorized request")
	}
}
