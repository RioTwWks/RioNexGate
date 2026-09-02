package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"rionexgate/internal/config"
)

func TestStealthDefaults(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := `server:
  api_key: test
database:
  path: ./data/test.db
core:
  stealth:
    enabled: true
    reality:
      dest: "cdn.example.com:443"
      server_names: ["cdn.example.com"]
      private_key: priv
      public_key: pub
      short_ids: ["abcd"]
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONFIG_PATH", cfgPath)

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	s := cfg.Core.Stealth
	if s.FingerprintOrDefault() != "firefox" {
		t.Fatalf("expected default fingerprint firefox, got %s", s.Fingerprint)
	}
	if s.XHTTP.Mode != "stream-one" {
		t.Fatalf("expected stream-one mode, got %s", s.XHTTP.Mode)
	}
	if s.XHTTP.Port != 443 || s.Vision.Port != 8443 {
		t.Fatalf("unexpected ports: xhttp=%d vision=%d", s.XHTTP.Port, s.Vision.Port)
	}
	if !s.IsActive() {
		t.Fatal("expected active stealth config")
	}
}
