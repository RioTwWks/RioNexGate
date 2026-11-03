package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

// --- publicBase / host-header injection (issue #43) ---

func setupPublicBaseRouter(t *testing.T, serverExtra string) http.Handler {
	t.Helper()
	if testing.Short() {
		t.Skip("sqlite")
	}
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := `server:
  port: 8080
  api_key: "test-secret-key16"
` + serverExtra + `database:
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
	return api.NewRouter(cfg, database, core.NewManager(cfg, database))
}

func createUserForLinkTest(t *testing.T, router http.Handler) uint {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/users",
		bytes.NewReader([]byte(`{"email":"link@example.com","traffic_gb":1,"expire_days":1}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-secret-key16")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create user: %d body=%s", rr.Code, rr.Body.String())
	}
	var u struct {
		ID uint `json:"id"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&u); err != nil {
		t.Fatal(err)
	}
	return u.ID
}

func mintLink(t *testing.T, router http.Handler, userID uint, host string) (int, string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost,
		"/api/users/"+strconv.FormatUint(uint64(userID), 10)+"/subscription-link",
		bytes.NewReader([]byte(`{}`)))
	req.Host = host
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-secret-key16")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	var resp struct {
		URL string `json:"url"`
	}
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	return rr.Code, resp.URL
}

func TestPublicBasePrefersConfiguredURL(t *testing.T) {
	router := setupPublicBaseRouter(t, `  public_base_url: "https://panel.example.com"
`)
	id := createUserForLinkTest(t, router)

	code, url := mintLink(t, router, id, "evil.example.com")
	if code != http.StatusOK {
		t.Fatalf("mint: %d", code)
	}
	if !strings.HasPrefix(url, "https://panel.example.com/api/subscription/") {
		t.Fatalf("expected configured base URL to win over Host header, got %q", url)
	}
}

func TestPublicBaseAllowedHosts(t *testing.T) {
	router := setupPublicBaseRouter(t, `  allowed_hosts: ["panel.example.com"]
`)
	id := createUserForLinkTest(t, router)

	code, url := mintLink(t, router, id, "panel.example.com")
	if code != http.StatusOK {
		t.Fatalf("whitelisted host: %d", code)
	}
	if !strings.HasPrefix(url, "http://panel.example.com/api/subscription/") {
		t.Fatalf("expected header-derived URL for whitelisted host, got %q", url)
	}

	code, _ = mintLink(t, router, id, "evil.example.com")
	if code != http.StatusInternalServerError {
		t.Fatalf("non-whitelisted host: expected 500, got %d", code)
	}
}

func TestPublicBaseLegacyFallbackWithoutWhitelist(t *testing.T) {
	router := setupPublicBaseRouter(t, "")
	id := createUserForLinkTest(t, router)

	code, url := mintLink(t, router, id, "legacy.example.com")
	if code != http.StatusOK {
		t.Fatalf("legacy fallback: %d", code)
	}
	if !strings.HasPrefix(url, "http://legacy.example.com/api/subscription/") {
		t.Fatalf("expected Host-derived URL in legacy mode, got %q", url)
	}
}
