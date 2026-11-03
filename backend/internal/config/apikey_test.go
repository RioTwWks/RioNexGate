package config

import (
	"os"
	"testing"
)

func TestValidateAPIKey(t *testing.T) {
	t.Setenv("RIONEXGATE_ALLOW_INSECURE_API_KEY", "")
	if err := ValidateAPIKey(""); err == nil {
		t.Fatal("expected error for empty key")
	}
	if err := ValidateAPIKey("change-me-to-secure-key"); err == nil {
		t.Fatal("expected error for placeholder key")
	}
	if err := ValidateAPIKey("short"); err == nil {
		t.Fatal("expected error for short key")
	}
	if err := ValidateAPIKey("long-enough-api-key"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Setenv("RIONEXGATE_ALLOW_INSECURE_API_KEY", "1")
	if err := ValidateAPIKey("change-me-to-secure-key"); err != nil {
		t.Fatalf("bypass should allow placeholder: %v", err)
	}
}

func TestValidateSkadiToken(t *testing.T) {
	t.Setenv("RIONEXGATE_ALLOW_INSECURE_API_KEY", "")
	for _, tok := range []string{"", "change-me-skadi-api-token", "rionexgate-skadi-api", "short"} {
		if err := ValidateSkadiToken(tok); err == nil {
			t.Fatalf("expected error for insecure token %q", tok)
		}
	}
	if err := ValidateSkadiToken("long-enough-skadi-token"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Setenv("RIONEXGATE_ALLOW_INSECURE_API_KEY", "1")
	if err := ValidateSkadiToken("change-me-skadi-api-token"); err != nil {
		t.Fatalf("bypass should allow placeholder: %v", err)
	}
}

func TestLoadRejectsInsecureSkadiToken(t *testing.T) {
	t.Setenv("RIONEXGATE_ALLOW_INSECURE_API_KEY", "")
	dir := t.TempDir()
	cfgPath := dir + "/config.yaml"
	yaml := `
server:
  api_key: "long-enough-api-key"
database:
  path: "` + dir + `/test.db"
core:
  type: "skadi"
  skadi:
    api_token: "change-me-skadi-api-token"
`
	if err := os.WriteFile(cfgPath, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONFIG_PATH", cfgPath)
	if _, err := Load(); err == nil {
		t.Fatal("expected Load to reject placeholder skadi api_token")
	}
}
