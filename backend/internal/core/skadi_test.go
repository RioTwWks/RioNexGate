package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"rionexgate/internal/config"
	"rionexgate/internal/models"
)

func TestNormalizeCoreType(t *testing.T) {
	cases := map[string]string{
		"xray":       "xray",
		"sing-box":   "sing-box",
		"singbox":    "sing-box",
		"skadi":      "skadi",
		"SkadiCore":  "skadi",
		"skadi-core": "skadi",
	}
	for in, want := range cases {
		got, err := NormalizeCoreType(in)
		if err != nil {
			t.Fatalf("%s: %v", in, err)
		}
		if got != want {
			t.Fatalf("%s: got %q want %q", in, got, want)
		}
	}
	if _, err := NormalizeCoreType("unknown"); err == nil {
		t.Fatal("expected error for unknown core")
	}
}

func TestGenerateSkadiConfigBasic(t *testing.T) {
	users := []models.User{
		{UUID: "b831381d-6324-4d53-ad4f-8cda48b30811", Email: "a@example.com"},
	}
	data, err := generateSkadiConfig(443, config.SkadiConfig{}, users, nil)
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	for _, want := range []string{
		`listen = ["0.0.0.0:443", "[::]:443"]`,
		`id = "b831381d-6324-4d53-ad4f-8cda48b30811"`,
		`email = "a@example.com"`,
		`[protocol.vless]`,
		`enabled = true`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in:\n%s", want, body)
		}
	}
	if strings.Contains(body, "[transport.reality]") {
		t.Fatalf("unexpected reality without stealth:\n%s", body)
	}
}

func TestGenerateSkadiConfigStealth(t *testing.T) {
	users := []models.User{
		{UUID: "b831381d-6324-4d53-ad4f-8cda48b30811", Email: "a@example.com"},
	}
	stealth := testStealthConfig()
	cfg := config.SkadiConfig{
		APIAddress:     "host.docker.internal:10086",
		APIToken:       "secret-token",
		MetricsAddress: "127.0.0.1:9091",
	}
	data, err := generateSkadiConfig(443, cfg, users, stealth)
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	checks := []string{
		`listen = ["0.0.0.0:443", "[::]:443"]`,
		`[transport.reality]`,
		`dest = "www.microsoft.com:443"`,
		`private_key = "SNX6hIY7eBmqDCdiR9HhycMkyuKtRty3PqJnhgAsn3w"`,
		`[transport.xhttp]`,
		`path = "/api/v1/data"`,
		`mode = "stream-one"`,
		`listen = "127.0.0.1:10086"`,
		`token = "secret-token"`,
		`listen = "127.0.0.1:9091"`,
	}
	for _, want := range checks {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in:\n%s", want, body)
		}
	}
	if strings.Contains(body, "xtls-rprx-vision") {
		t.Fatalf("skadi config must not emit vision flow:\n%s", body)
	}
}

func TestAdaptProfilesForSkadi(t *testing.T) {
	user := models.User{UUID: "b831381d-6324-4d53-ad4f-8cda48b30811", Email: "u@test.com"}
	stealth := testStealthConfig()
	xrayProfiles := GetClientLinkProfiles("host.example", 443, user, stealth, nil, nil, nil, "")
	if len(xrayProfiles) < 2 {
		t.Fatalf("expected xray stealth profiles, got %d", len(xrayProfiles))
	}
	skadi := adaptProfilesForSkadi("host.example", 443, user, stealth, xrayProfiles)
	if len(skadi) < 2 {
		t.Fatalf("expected skadi profiles, got %+v", skadi)
	}
	for _, p := range skadi {
		if strings.Contains(p.Link, "xtls-rprx-vision") {
			t.Fatalf("vision flow must not appear for skadi: %s", p.Link)
		}
		if p.Profile == "vision-tcp-primary" {
			t.Fatalf("vision profile must be omitted: %+v", p)
		}
	}
	if skadi[0].Transport != "xhttp" {
		t.Fatalf("expected xhttp primary, got %+v", skadi[0])
	}
	if skadi[1].Transport != "tcp" || !strings.Contains(skadi[1].Link, "security=reality") {
		t.Fatalf("expected reality tcp secondary, got %+v", skadi[1])
	}
}

func TestSkadiAPIListenAddress(t *testing.T) {
	if got := skadiAPIListenAddress("host.docker.internal:10086"); got != "127.0.0.1:10086" {
		t.Fatalf("got %q", got)
	}
	if got := skadiAPIListenAddress("127.0.0.1:10086"); got != "127.0.0.1:10086" {
		t.Fatalf("got %q", got)
	}
}

// testStealthConfigSkadiValid uses REALITY keys that pass skadicore check-config (v0.1.3+).
func testStealthConfigSkadiValid() *config.StealthConfig {
	s := testStealthConfig()
	s.Reality.PrivateKey = "l8KJbSSnlO29SIz+rkfBWiI04ax+pMrvKT10LezmIDA="
	s.Reality.PublicKey = "mTQRWF244Qsvn9UuDjOBlVPBTYZIWOUx6LwS3OeM93I"
	s.Reality.ShortIDs = []string{"86c10d238f7ac5c0"}
	return s
}

func skadicoreBinary() string {
	if p := strings.TrimSpace(os.Getenv("SKADICORE_BIN")); p != "" {
		return p
	}
	p, err := exec.LookPath("skadicore")
	if err != nil {
		return ""
	}
	return p
}

func TestSkadiConfigValidate(t *testing.T) {
	if os.Getenv("CI") == "" && os.Getenv("RUN_SKADI_TEST") == "" {
		t.Skip("set RUN_SKADI_TEST=1 or run in CI to validate with skadicore binary")
	}
	bin := skadicoreBinary()
	if bin == "" {
		t.Skip("skadicore binary not in PATH (set SKADICORE_BIN)")
	}

	users := []models.User{{UUID: "b831381d-6324-4d53-ad4f-8cda48b30811", Email: "test@example.com"}}
	cfg := config.SkadiConfig{
		APIAddress:     "127.0.0.1:10086",
		APIToken:       "integration-test-token",
		MetricsAddress: "127.0.0.1:9091",
	}
	data, err := generateSkadiConfig(443, cfg, users, testStealthConfigSkadiValid())
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "skadi.toml")
	if err := os.WriteFile(cfgPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(bin, "check-config", "--config", cfgPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("skadicore check-config failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "Configuration OK") {
		t.Fatalf("unexpected check-config output:\n%s", out)
	}
}

func TestBuildClientConfigSkadi(t *testing.T) {
	user := models.User{UUID: "b831381d-6324-4d53-ad4f-8cda48b30811", Email: "u@test.com"}
	cfg, err := BuildClientConfig("host.example", 443, user, 10808, testStealthConfig(), nil, nil, nil, nil, "skadi")
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Servers) != 1 || cfg.Servers[0].Protocol != "vless" {
		t.Fatalf("expected single vless server: %+v", cfg.Servers)
	}
	if strings.Contains(cfg.Servers[0].Link, "xtls-rprx-vision") {
		t.Fatalf("skadi client link must not use vision: %s", cfg.Servers[0].Link)
	}
	for _, p := range cfg.Profiles {
		if strings.Contains(p.Link, "xtls-rprx-vision") {
			t.Fatalf("vision in skadi profile: %s", p.Link)
		}
	}
}
