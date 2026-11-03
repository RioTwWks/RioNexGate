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

func TestDocsDefaultDisabled(t *testing.T) {
	c := &Config{}
	if c.DocsEnabled() {
		t.Fatal("nil EnableDocs must default to false")
	}
	tru := true
	c.Server.EnableDocs = &tru
	if !c.DocsEnabled() {
		t.Fatal("explicit enable_docs=true must be honored")
	}
}
