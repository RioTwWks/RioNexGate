package core

import (
	"strings"
	"testing"

	"rionexgate/internal/models"
)

func TestBuildSubscriptionBase64(t *testing.T) {
	user := models.User{
		UUID:  "550e8400-e29b-41d4-a716-446655440000",
		Email: "test@example.com",
	}
	links := BuildSubscriptionLinks("example.com", 443, user, nil, nil)
	if !strings.Contains(strings.Join(links, "\n"), "vless://") {
		t.Fatalf("expected vless link in subscription")
	}
}

func TestBuildClientConfigHash(t *testing.T) {
	user := models.User{
		UUID:  "550e8400-e29b-41d4-a716-446655440000",
		Email: "test@example.com",
	}
	cfg, err := BuildClientConfig("example.com", 443, user, 10808, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ConfigHash == "" || len(cfg.Servers) == 0 {
		t.Fatalf("invalid config: %+v", cfg)
	}
	if len(cfg.Profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(cfg.Profiles))
	}
	if cfg.Profiles[0].Priority != 1 || cfg.Profiles[0].Transport != "tcp" {
		t.Fatalf("unexpected profile: %+v", cfg.Profiles[0])
	}
}

func TestBuildClientConfigStealthProfiles(t *testing.T) {
	user := models.User{UUID: "uuid-1", Email: "user@test.com"}
	cfg, err := BuildClientConfig("host.example", 443, user, 10808, testStealthConfig(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(cfg.Profiles))
	}
	if cfg.Profiles[0].Profile != "xhttp-primary" {
		t.Fatalf("unexpected primary profile: %+v", cfg.Profiles[0])
	}
}
