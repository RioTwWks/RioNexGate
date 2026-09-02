package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"rionexgate/internal/db"
)

func TestDeviceTokenAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("sqlite")
	}

	dir := t.TempDir()
	database, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(); err != nil {
		t.Fatal(err)
	}

	user, err := database.CreateUser(db.CreateUserInput{Email: "dev@test.com", TrafficGB: 1, ExpireDays: 7})
	if err != nil {
		t.Fatal(err)
	}
	device, err := database.CreateDevice(user.ID, "test")
	if err != nil {
		t.Fatal(err)
	}

	var gotUserID uint
	handler := DeviceTokenAuth(database)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dc, ok := DeviceFromContext(r.Context())
		if !ok {
			t.Fatal("missing device context")
		}
		gotUserID = dc.User.ID
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.Background())
	req.Header.Set("X-Device-Token", device.Token)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || gotUserID != user.ID {
		t.Fatalf("expected authorized device, code=%d user=%d", rr.Code, gotUserID)
	}
}
