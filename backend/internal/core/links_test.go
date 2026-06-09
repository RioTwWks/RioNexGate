package core

import (
	"strings"
	"testing"
	"time"

	"proxy-mgr/internal/models"
)

func TestBuildVLESSLink(t *testing.T) {
	user := models.User{
		UUID:  "550e8400-e29b-41d4-a716-446655440000",
		Email: "test@example.com",
	}
	link := GetClientLink("example.com", 443, user, "vless")
	if !strings.HasPrefix(link, "vless://") {
		t.Fatalf("expected vless prefix, got %s", link)
	}
	if !strings.Contains(link, user.UUID) {
		t.Fatalf("expected uuid in link")
	}
	if !strings.Contains(link, "example.com:443") {
		t.Fatalf("expected host:port in link")
	}
	_ = time.Now()
}

func TestGenerateXrayConfig(t *testing.T) {
	users := []models.User{
		{UUID: "uuid-1", Email: "a@example.com"},
		{UUID: "uuid-2", Email: "b@example.com"},
	}
	data, err := generateXrayConfig(443, "127.0.0.1:10085", users)
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	if !strings.Contains(body, "uuid-1") || !strings.Contains(body, "a@example.com") {
		t.Fatalf("missing user in config: %s", body)
	}
}
