package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"rionexgate/internal/config"
)

func TestClientRegisterAuth(t *testing.T) {
	cfg := &config.Config{Server: config.ServerConfig{
		APIKey:             "admin-api-key-16",
		RegistrationSecret: "reg-secret-value1",
	}}
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	h := ClientRegisterAuth(cfg)(okHandler)

	req := httptest.NewRequest(http.MethodPost, "/api/client/register", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without credentials, got %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/client/register", nil)
	req.Header.Set("X-API-Key", "admin-api-key-16")
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 with API key, got %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/client/register", nil)
	req.Header.Set("X-Registration-Secret", "reg-secret-value1")
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 with registration secret, got %d", rr.Code)
	}

	open := &config.Config{Server: config.ServerConfig{APIKey: "admin-api-key-16", AllowOpenRegister: true}}
	hOpen := ClientRegisterAuth(open)(okHandler)
	req = httptest.NewRequest(http.MethodPost, "/api/client/register", nil)
	rr = httptest.NewRecorder()
	hOpen.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 with allow_open_register, got %d", rr.Code)
	}
}
