package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"proxy-mgr/internal/api"
	"proxy-mgr/internal/config"
	"proxy-mgr/internal/core"
	"proxy-mgr/internal/db"
)

func setupTestServer(t *testing.T) (*httptest.Server, *db.DB) {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	xrayPath := filepath.Join(dir, "xray", "config.json")

	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080, APIKey: "test-secret"},
		Database: config.DatabaseConfig{Path: dbPath},
		Core: config.CoreConfig{
			Type:       "xray",
			ListenPort: 443,
			PublicHost: "test.local",
			Xray: config.XrayConfig{
				ConfigPath: xrayPath,
				APIAddress: "127.0.0.1:10085",
			},
			StatsPoll: 60,
		},
		Limits: config.LimitsConfig{DefaultTrafficGB: 10, DefaultExpireDays: 7},
	}

	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(); err != nil {
		t.Fatal(err)
	}

	coreMgr := core.NewManager(cfg, database)
	router := api.NewRouter(cfg, database, coreMgr)
	return httptest.NewServer(router), database
}

func TestAPIE2E_UserCRUDAndLinks(t *testing.T) {
	if os.Getenv("CGO_ENABLED") == "0" {
		t.Skip("sqlite requires CGO")
	}

	srv, _ := setupTestServer(t)
	defer srv.Close()

	client := &http.Client{}
	base := srv.URL + "/api"
	authHeader := func(req *http.Request) {
		req.Header.Set("X-API-Key", "test-secret")
	}

	// health (no auth)
	resp, err := http.Get(base + "/health")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health: expected 200, got %d", resp.StatusCode)
	}

	// openapi + docs
	for _, path := range []string{"/openapi.yaml", "/docs"} {
		resp, err = http.Get(base + path)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d", path, resp.StatusCode)
		}
	}

	// protocols
	req, _ := http.NewRequest(http.MethodGet, base+"/protocols", nil)
	authHeader(req)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var protoResp struct {
		Protocols []string `json:"protocols"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&protoResp); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if len(protoResp.Protocols) < 3 {
		t.Fatalf("expected protocols list, got %+v", protoResp.Protocols)
	}

	// create user
	body := []byte(`{"email":"e2e@example.com","traffic_gb":5,"expire_days":14}`)
	req, _ = http.NewRequest(http.MethodPost, base+"/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	authHeader(req)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create user: expected 201, got %d", resp.StatusCode)
	}
	var user struct {
		ID    uint   `json:"id"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	// link variants
	for _, proto := range []string{"vless", "vmess", "trojan"} {
		req, _ = http.NewRequest(http.MethodGet, base+"/users/"+strconv.FormatUint(uint64(user.ID), 10)+"/link?proto="+proto, nil)
		authHeader(req)
		resp, err = client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		var linkResp struct {
			Link     string `json:"link"`
			Protocol string `json:"protocol"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&linkResp); err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if linkResp.Protocol != proto {
			t.Fatalf("expected protocol %s, got %s", proto, linkResp.Protocol)
		}
		if linkResp.Link == "" {
			t.Fatalf("empty link for %s", proto)
		}
	}

	// delete user
	req, _ = http.NewRequest(http.MethodDelete, base+"/users/"+strconv.FormatUint(uint64(user.ID), 10), nil)
	authHeader(req)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d", resp.StatusCode)
	}
}
