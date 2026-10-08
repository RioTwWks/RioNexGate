package middleware

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"rionexgate/internal/config"
	"rionexgate/internal/db"
)

func TestResolveRegisterAuth(t *testing.T) {
	cfg := &config.Config{Server: config.ServerConfig{
		APIKey:             "admin-api-key-16",
		RegistrationSecret: "reg-secret-value1",
	}}
	if auth := ResolveRegisterAuth(cfg, nil, "", "", ""); auth != nil {
		t.Fatal("expected nil without credentials")
	}
	if auth := ResolveRegisterAuth(cfg, nil, "admin-api-key-16", "", ""); auth == nil || auth.Mode != "api_key" {
		t.Fatalf("api key: %+v", auth)
	}
	if auth := ResolveRegisterAuth(cfg, nil, "", "reg-secret-value1", ""); auth == nil || auth.Mode != "registration_secret" {
		t.Fatalf("reg secret: %+v", auth)
	}
	open := &config.Config{Server: config.ServerConfig{APIKey: "admin-api-key-16", AllowOpenRegister: true}}
	if auth := ResolveRegisterAuth(open, nil, "", "", ""); auth == nil || auth.Mode != "open" {
		t.Fatalf("open: %+v", auth)
	}
}

func TestClientRegisterAuthInviteHeader(t *testing.T) {
	if testing.Short() {
		t.Skip("sqlite")
	}
	database, err := db.Open(filepath.Join(t.TempDir(), "reg.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(); err != nil {
		t.Fatal(err)
	}
	user, err := database.CreateUser(db.CreateUserInput{Email: "a@example.com", TrafficGB: 1, ExpireDays: 7})
	if err != nil {
		t.Fatal(err)
	}
	inv, err := database.CreateInvite(db.CreateInviteInput{UserID: user.ID, MaxUses: 1, ExpiresIn: time.Hour})
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{Server: config.ServerConfig{APIKey: "admin-api-key-16"}}
	var mode string
	h := ClientRegisterAuth(cfg, database)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth, ok := RegisterAuthFromContext(r.Context()); ok {
			mode = auth.Mode
		}
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/client/register", nil)
	req.Header.Set("X-Invite-Token", inv.Token)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated || mode != "invite" {
		t.Fatalf("code=%d mode=%s", rr.Code, mode)
	}
}
