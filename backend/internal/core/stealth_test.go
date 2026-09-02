package core

import (
	"encoding/json"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"rionexgate/internal/config"
	"rionexgate/internal/models"
)

func testStealthConfig() *config.StealthConfig {
	return &config.StealthConfig{
		Enabled:     true,
		Fingerprint: "firefox",
		Reality: config.StealthRealityConfig{
			Dest:        "www.microsoft.com:443",
			ServerNames: []string{"www.microsoft.com"},
			PrivateKey:  "SNX6hIY7eBmqDCdiR9HhycMkyuKtRty3PqJnhgAsn3w",
			PublicKey:   "Izo7I-b-XfLZP0jTBHhC3zzHZ02-oX57Z1JwB6fgABM",
			ShortIDs:    []string{"a1b2c3d4"},
			Show:        false,
			Xver:        0,
		},
		XHTTP: config.StealthXHTTPConfig{
			Enabled: true,
			Port:    443,
			Path:    "/api/v1/data",
			Mode:    "stream-one",
			Tag:     "vless-xhttp-reality",
		},
		Vision: config.StealthVisionConfig{
			Enabled: true,
			Port:    8443,
			Tag:     "vless-vision-reality",
		},
		TLS: config.StealthTLSConfig{
			Enabled: false,
		},
	}
}

func TestStealthConfigIsActive(t *testing.T) {
	s := testStealthConfig()
	if !s.IsActive() {
		t.Fatal("expected stealth config to be active")
	}
	s.Reality.PublicKey = ""
	if s.IsActive() {
		t.Fatal("expected inactive without public key")
	}
}

func TestGenerateStealthXrayConfig(t *testing.T) {
	users := []models.User{
		{UUID: "uuid-1", Email: "a@example.com"},
	}
	stealth := testStealthConfig()
	data, err := generateXrayConfig(443, "127.0.0.1:10085", users, stealth, MultihopData{})
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)

	checks := []string{
		`"network": "xhttp"`,
		`"mode": "stream-one"`,
		`"dest": "www.microsoft.com:443"`,
		`"privateKey": "SNX6hIY7eBmqDCdiR9HhycMkyuKtRty3PqJnhgAsn3w"`,
		`"shortIds": ["a1b2c3d4"]`,
		`"flow": "xtls-rprx-vision"`,
		`"port": 443`,
		`"port": 8443`,
		`vless-xhttp-reality`,
		`vless-vision-reality`,
	}
	for _, want := range checks {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in config:\n%s", want, body)
		}
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	inbounds, ok := parsed["inbounds"].([]interface{})
	if !ok || len(inbounds) != 2 {
		t.Fatalf("expected 2 inbounds, got %d", len(inbounds))
	}
}

func TestBuildVLESSRealityXHTTPLink(t *testing.T) {
	user := models.User{UUID: "550e8400-e29b-41d4-a716-446655440000", Email: "test@example.com"}
	stealth := testStealthConfig()
	link := buildVLESSRealityXHTTPLink("proxy.example.com", 443, user, stealth)

	u, err := url.Parse(link)
	if err != nil {
		t.Fatal(err)
	}
	if u.Scheme != "vless" {
		t.Fatalf("expected vless scheme, got %s", u.Scheme)
	}
	q := u.Query()
	expect := map[string]string{
		"encryption": "none",
		"type":       "xhttp",
		"security":   "reality",
		"sni":        "www.microsoft.com",
		"fp":         "firefox",
		"pbk":        "Izo7I-b-XfLZP0jTBHhC3zzHZ02-oX57Z1JwB6fgABM",
		"sid":        "a1b2c3d4",
		"path":       "/api/v1/data",
		"mode":       "stream-one",
	}
	for k, v := range expect {
		if q.Get(k) != v {
			t.Fatalf("param %s: want %q, got %q", k, v, q.Get(k))
		}
	}
}

func TestBuildVLESSRealityVisionLink(t *testing.T) {
	user := models.User{UUID: "550e8400-e29b-41d4-a716-446655440000", Email: "test@example.com"}
	stealth := testStealthConfig()
	link := buildVLESSRealityVisionLink("proxy.example.com", 8443, user, stealth)

	u, err := url.Parse(link)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if q.Get("flow") != "xtls-rprx-vision" {
		t.Fatalf("expected vision flow, got %q", q.Get("flow"))
	}
	if q.Get("security") != "reality" {
		t.Fatalf("expected reality security")
	}
	if q.Get("type") != "tcp" {
		t.Fatalf("expected tcp type")
	}
}

func TestBuildVLESSTLSLink(t *testing.T) {
	user := models.User{UUID: "550e8400-e29b-41d4-a716-446655440000", Email: "test@example.com"}
	stealth := testStealthConfig()
	stealth.TLS = config.StealthTLSConfig{
		Enabled: true,
		Port:    2053,
		SNI:     "tls.example.com",
		ALPN:    []string{"h2", "http/1.1"},
	}
	link := buildVLESSTLSLink("proxy.example.com", 2053, user, stealth)

	u, err := url.Parse(link)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if q.Get("security") != "tls" {
		t.Fatalf("expected tls security")
	}
	if q.Get("sni") != "tls.example.com" {
		t.Fatalf("expected sni")
	}
	if q.Get("alpn") != "h2,http/1.1" {
		t.Fatalf("expected alpn, got %q", q.Get("alpn"))
	}
}

func TestGetClientLinkProfiles(t *testing.T) {
	user := models.User{UUID: "uuid-1", Email: "user@test.com"}
	stealth := testStealthConfig()
	profiles := GetClientLinkProfiles("host.example", 443, user, stealth)
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}
	if profiles[0].Profile != "xhttp-primary" || profiles[0].Priority != 1 {
		t.Fatalf("unexpected primary profile: %+v", profiles[0])
	}
	if profiles[1].Profile != "vision-ios-fallback" {
		t.Fatalf("unexpected fallback profile: %+v", profiles[1])
	}
}

func TestGenerateStealthXrayConfigNoFragmentationOnReality(t *testing.T) {
	users := []models.User{{UUID: "uuid-1", Email: "a@example.com"}}
	stealth := testStealthConfig()
	stealth.Fragmentation = config.StealthFragmentationConfig{Enabled: true, Strategy: "serverhello"}
	data, err := generateXrayConfig(443, "127.0.0.1:10085", users, stealth, MultihopData{})
	if err != nil { t.Fatal(err) }
	if strings.Contains(string(data), `"finalmask"`) { t.Fatal("finalmask must not be on REALITY inbounds") }
}

func TestGenerateStealthXrayConfigFragmentationOnTLS(t *testing.T) {
	users := []models.User{{UUID: "uuid-1", Email: "a@example.com"}}
	stealth := testStealthConfig()
	stealth.TLS = config.StealthTLSConfig{Enabled: true, Port: 2053, SNI: "tls.example.com", Tag: "vless-tls"}
	stealth.Fragmentation = config.StealthFragmentationConfig{Enabled: true, Strategy: "serverhello", Length: "60-120"}
	data, err := generateXrayConfig(443, "127.0.0.1:10085", users, stealth, MultihopData{})
	if err != nil { t.Fatal(err) }
	if !strings.Contains(string(data), `"packets": "tlshello"`) { t.Fatal("expected tlshello fragmentation") }
}


func TestEncodeSubscription(t *testing.T) {
	user := models.User{UUID: "uuid-1", Email: "user@test.com"}
	profiles := GetClientLinkProfiles("host.example", 443, user, testStealthConfig())
	encoded := EncodeSubscription(profiles)
	if encoded == "" {
		t.Fatal("empty subscription")
	}
	decoded := FormatSubscriptionLinks(profiles)
	if !strings.Contains(decoded, "vless://") {
		t.Fatalf("expected vless links in subscription")
	}
	lines := strings.Split(strings.TrimSpace(decoded), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 subscription lines, got %d", len(lines))
	}
}

func TestXrayConfigValidate(t *testing.T) {
	if os.Getenv("CI") == "" && os.Getenv("RUN_XRAY_TEST") == "" {
		t.Skip("set RUN_XRAY_TEST=1 or run in CI to validate with xray binary")
	}
	if _, err := exec.LookPath("xray"); err != nil {
		t.Skip("xray binary not in PATH")
	}

	users := []models.User{{UUID: "550e8400-e29b-41d4-a716-446655440000", Email: "test@example.com"}}
	data, err := generateXrayConfig(443, "127.0.0.1:10085", users, testStealthConfig(), MultihopData{})
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(cfgPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("xray", "run", "-test", "-c", cfgPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("xray run -test failed: %v\n%s", err, out)
	}
}
