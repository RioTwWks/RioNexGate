package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"rionexgate/internal/api"
	"rionexgate/internal/config"
	"rionexgate/internal/core"
	"rionexgate/internal/db"
)

func TestRotateAPIKeyHotSwap(t *testing.T) {
	if testing.Short() {
		t.Skip("sqlite")
	}
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := `server:
  port: 8080
  api_key: "initial-api-key16"
database:
  path: ` + filepath.Join(dir, "t.db") + `
core:
  type: xray
  listen_port: 443
  public_host: test.local
  xray:
    config_path: ` + filepath.Join(dir, "xray.json") + `
    api_address: "127.0.0.1:10085"
limits:
  default_traffic_gb: 10
  default_expire_days: 7
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONFIG_PATH", cfgPath)
	t.Setenv("RIONEXGATE_ALLOW_INSECURE_API_KEY", "1")

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	database, err := db.Open(cfg.Database.Path)
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(); err != nil {
		t.Fatal(err)
	}
	coreMgr := core.NewManager(cfg, database)
	router := api.NewRouter(cfg, database, coreMgr)

	// Old key works.
	req := httptest.NewRequest(http.MethodGet, "/api/protocols", nil)
	req.Header.Set("X-API-Key", "initial-api-key16")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("old key: %d", rr.Code)
	}

	body := []byte(`{"new_key":"rotated-api-key16"}`)
	req = httptest.NewRequest(http.MethodPost, "/api/security/rotate-api-key", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "initial-api-key16")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("rotate: %d body=%s", rr.Code, rr.Body.String())
	}
	var resp struct {
		APIKey string `json:"api_key"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.APIKey != "rotated-api-key16" {
		t.Fatalf("got %q", resp.APIKey)
	}

	// Old key rejected; new key works.
	req = httptest.NewRequest(http.MethodGet, "/api/protocols", nil)
	req.Header.Set("X-API-Key", "initial-api-key16")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("old key after rotate: %d", rr.Code)
	}
	req = httptest.NewRequest(http.MethodGet, "/api/protocols", nil)
	req.Header.Set("X-API-Key", "rotated-api-key16")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("new key: %d", rr.Code)
	}
}
