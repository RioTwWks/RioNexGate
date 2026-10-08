package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPersistAPIKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `server:
  port: 8080
  api_key: "old-key-value-16ch"   # comment
database:
  path: ./data/x.db
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	newKey := "brand-new-api-key1"
	if err := PersistAPIKey(path, newKey); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	if !strings.Contains(body, `api_key: "`+newKey+`"`) {
		t.Fatalf("key not updated: %s", body)
	}
	if !strings.Contains(body, "# comment") {
		t.Fatalf("comment lost: %s", body)
	}
	if strings.Contains(body, "old-key-value-16ch") {
		t.Fatal("old key still present")
	}
}

func TestGenerateAPIKey(t *testing.T) {
	k, err := GenerateAPIKey()
	if err != nil {
		t.Fatal(err)
	}
	if len(k) != 64 {
		t.Fatalf("len=%d", len(k))
	}
	if err := ValidateAPIKey(k); err != nil {
		t.Fatal(err)
	}
}
