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
	links := BuildSubscriptionLinks("example.com", 443, user, nil, nil, nil, nil, nil)
	if !strings.Contains(strings.Join(links, "\n"), "vless://") {
		t.Fatalf("expected vless link in subscription")
	}
}

func TestBuildClientConfigHash(t *testing.T) {
	user := models.User{
		UUID:  "550e8400-e29b-41d4-a716-446655440000",
		Email: "test@example.com",
	}
	cfg, err := BuildClientConfig("example.com", 443, user, 10808, nil, nil, nil, nil, nil)
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
	cfg, err := BuildClientConfig("host.example", 443, user, 10808, testStealthConfig(), nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if p := vlessLinkPort(cfg.Servers[0].Link, 443); p != 8443 {
		t.Fatalf("vlessLinkPort on real link: expected 8443, got %d link=%s", p, cfg.Servers[0].Link)
	}
}

func TestBuildClientConfigStealthProfiles(t *testing.T) {
	user := models.User{UUID: "uuid-1", Email: "user@test.com"}
	cfg, err := BuildClientConfig("host.example", 443, user, 10808, testStealthConfig(), nil, nil, nil, nil)
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

func TestBuildClientConfigStealthOmitsLegacyServers(t *testing.T) {
	user := models.User{UUID: "uuid-1", Email: "user@test.com"}
	cfg, err := BuildClientConfig("host.example", 443, user, 10808, testStealthConfig(), nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Servers) != 1 {
		t.Fatalf("expected 1 server (vless only), got %d", len(cfg.Servers))
	}
	if cfg.Servers[0].Protocol != "vless" {
		t.Fatalf("expected vless server only, got %+v", cfg.Servers[0])
	}
}

func TestVlessTransportParamsIncludesReality(t *testing.T) {
	user := models.User{UUID: "uuid-1", Email: "user@test.com"}
	link := buildVLESSRealityVisionLink("host.example", 8443, user, testStealthConfig())
	params := vlessTransportParams(link)
	for _, key := range []string{"security", "flow", "sni", "fp", "pbk", "sid"} {
		if params[key] == "" {
			t.Fatalf("expected %s in transport params, got %+v", key, params)
		}
	}
	if params["security"] != "reality" || params["flow"] != "xtls-rprx-vision" {
		t.Fatalf("unexpected reality params: %+v", params)
	}
	if params["type"] != "tcp" {
		t.Fatalf("type must not include URL fragment, got %q", params["type"])
	}
}

func TestVlessTransportParamsStripsFragmentWithEmail(t *testing.T) {
	// type is the last query key alphabetically; without fragment stripping it becomes
	// "tcp#test@test.test-vision" and breaks RioNexTunnel transport parsing.
	user := models.User{UUID: "550e8400-e29b-41d4-a716-446655440000", Email: "test@test.test"}
	stealth := testStealthConfig()
	stealth.Reality.ServerNames = []string{"www.bol.com"}
	stealth.Reality.Dest = "www.bol.com:443"
	link := buildVLESSRealityVisionLink("rio2skadi.ru", 8443, user, stealth)
	if !strings.Contains(link, "#test@test.test-vision") {
		t.Fatalf("expected vision fragment in link, got %s", link)
	}
	params := vlessTransportParams(link)
	if params["type"] != "tcp" {
		t.Fatalf("expected type=tcp, got %q (link ends with fragment after type param)", params["type"])
	}
	if params["security"] != "reality" || params["flow"] != "xtls-rprx-vision" {
		t.Fatalf("unexpected reality params: %+v", params)
	}
	cfg, err := BuildClientConfig("rio2skadi.ru", 443, user, 10808, stealth, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(cfg.Servers))
	}
	if cfg.Servers[0].Params["type"] != "tcp" {
		t.Fatalf("BuildClientConfig transport type: got %q", cfg.Servers[0].Params["type"])
	}
	if cfg.Servers[0].Port != 8443 {
		t.Fatalf("expected vision port 8443, got %d", cfg.Servers[0].Port)
	}
	if cfg.Inbounds.SOCKS5.Port != 10808 || cfg.Inbounds.SOCKS5.Auth != "none" {
		t.Fatalf("unexpected inbounds: %+v", cfg.Inbounds)
	}
	if len(cfg.DNS.Servers) != 2 {
		t.Fatalf("unexpected dns: %+v", cfg.DNS)
	}
	if len(cfg.Profiles) != 2 {
		t.Fatalf("expected 2 profiles (vision+xhttp), got %d", len(cfg.Profiles))
	}
	if cfg.Profiles[0].Profile != "vision-tcp-primary" || cfg.Profiles[0].Transport != "tcp" || cfg.Profiles[0].Port != 8443 {
		t.Fatalf("unexpected primary profile: %+v", cfg.Profiles[0])
	}
	if cfg.Profiles[1].Profile != "xhttp-anti-dpi" || cfg.Profiles[1].Transport != "xhttp" {
		t.Fatalf("unexpected xhttp profile: %+v", cfg.Profiles[1])
	}
	if cfg.ConfigHash == "" {
		t.Fatal("expected non-empty config_hash")
	}
}
