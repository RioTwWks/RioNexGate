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
	cfg, err := BuildClientConfig("example.com", 443, user, 10808, nil, nil, nil)
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

func TestVlessLinkPort(t *testing.T) {
	link := "vless://uuid@host.example:8443?encryption=none&type=tcp&flow=xtls-rprx-vision"
	if p := vlessLinkPort(link, 443); p != 8443 {
		t.Fatalf("expected 8443, got %d", p)
	}
	user := models.User{UUID: "uuid-1", Email: "user@test.com"}
	cfg, err := BuildClientConfig("host.example", 443, user, 10808, testStealthConfig(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if p := vlessLinkPort(cfg.Servers[0].Link, 443); p != 8443 {
		t.Fatalf("vlessLinkPort on real link: expected 8443, got %d link=%s", p, cfg.Servers[0].Link)
	}
}

func TestBuildClientConfigStealthProfiles(t *testing.T) {
	user := models.User{UUID: "uuid-1", Email: "user@test.com"}
	cfg, err := BuildClientConfig("host.example", 443, user, 10808, testStealthConfig(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(cfg.Profiles))
	}
	if cfg.Profiles[0].Profile != "vision-tcp-primary" {
		t.Fatalf("unexpected primary profile: %+v", cfg.Profiles[0])
	}
	if cfg.Servers[0].Params["security"] != "reality" {
		t.Fatalf("expected reality security in vless server params, got %+v", cfg.Servers[0].Params)
	}
	if cfg.Servers[0].Port != 8443 {
		t.Fatalf("expected vision port 8443 in vless server, got %d (link=%s)", cfg.Servers[0].Port, cfg.Servers[0].Link)
	}
}
