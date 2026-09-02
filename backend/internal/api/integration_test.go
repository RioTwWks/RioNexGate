package api_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"rionexgate/internal/api"
	"rionexgate/internal/config"
	"rionexgate/internal/core"
	"rionexgate/internal/db"
	"rionexgate/internal/models"
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

func TestRioNexTunnelClientFlow(t *testing.T) {
	if os.Getenv("CGO_ENABLED") == "0" {
		t.Skip("sqlite requires CGO")
	}

	srv, database := setupTestServer(t)
	defer srv.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	base := srv.URL + "/api"
	authHeader := func(req *http.Request) {
		req.Header.Set("X-API-Key", "test-secret")
	}

	body := []byte(`{"email":"client@example.com","traffic_gb":5,"expire_days":14}`)
	req, _ := http.NewRequest(http.MethodPost, base+"/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	authHeader(req)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var user struct {
		ID uint `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	regBody := []byte(`{"user_id":` + strconv.FormatUint(uint64(user.ID), 10) + `,"label":"phone"}`)
	req, _ = http.NewRequest(http.MethodPost, base+"/client/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Version", "v1")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d", resp.StatusCode)
	}
	var regResp struct {
		DeviceToken     string `json:"device_token"`
		SubscriptionURL string `json:"subscription_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&regResp); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	deviceHeader := func(req *http.Request) {
		req.Header.Set("X-Device-Token", regResp.DeviceToken)
		req.Header.Set("X-API-Version", "v1")
	}

	req, _ = http.NewRequest(http.MethodGet, base+"/client/config", nil)
	deviceHeader(req)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var cfg1 struct {
		ConfigHash string `json:"config_hash"`
		Servers    []struct {
			Link string `json:"link"`
		} `json:"servers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&cfg1); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if cfg1.ConfigHash == "" || len(cfg1.Servers) == 0 {
		t.Fatalf("invalid config: %+v", cfg1)
	}

	statsBody := []byte(`{"session_id":"sess-1","bytes_in":1000,"bytes_out":2000,"status":"connected"}`)
	req, _ = http.NewRequest(http.MethodPost, base+"/client/stats", bytes.NewReader(statsBody))
	req.Header.Set("Content-Type", "application/json")
	deviceHeader(req)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	req, _ = http.NewRequest(http.MethodPost, base+"/client/stats", bytes.NewReader(statsBody))
	req.Header.Set("Content-Type", "application/json")
	deviceHeader(req)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	var count int64
	database.Model(&models.ClientStatsReport{}).Where("session_id = ?", "sess-1").Count(&count)
	if count != 1 {
		t.Fatalf("expected 1 stats report after dedup, got %d", count)
	}

	req, _ = http.NewRequest(http.MethodGet, base+"/client/config", nil)
	req.Header.Set("X-Device-Token", "invalid-token")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("invalid token: expected 401, got %d", resp.StatusCode)
	}

	u, err := database.GetUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	resp, err = http.Get(base + "/subscription/" + u.SubscriptionToken)
	if err != nil {
		t.Fatal(err)
	}
	subBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := base64.StdEncoding.DecodeString(string(subBody))
	if err != nil {
		t.Fatalf("subscription not valid base64: %v", err)
	}
	if !strings.Contains(string(decoded), "vless://") {
		t.Fatalf("subscription missing vless link: %s", decoded)
	}

	updateBody := []byte(`{"email":"client-updated@example.com"}`)
	req, _ = http.NewRequest(http.MethodPut, base+"/users/"+strconv.FormatUint(uint64(user.ID), 10), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	authHeader(req)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	req, _ = http.NewRequest(http.MethodGet, base+"/client/config", nil)
	deviceHeader(req)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var cfg2 struct {
		ConfigHash string `json:"config_hash"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&cfg2); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if cfg2.ConfigHash == cfg1.ConfigHash {
		t.Fatalf("expected config_hash to change after user update")
	}

	req, _ = http.NewRequest(http.MethodGet, base+"/client/commands?timeout=1", nil)
	deviceHeader(req)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("commands: expected 200, got %d", resp.StatusCode)
	}
}
