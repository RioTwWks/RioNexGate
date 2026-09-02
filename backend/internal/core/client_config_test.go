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
	links := BuildSubscriptionLinks("example.com", 443, user, nil, nil, nil)
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
}
