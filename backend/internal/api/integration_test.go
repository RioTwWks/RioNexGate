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
	return setupTestServerWith(t, nil)
}

func setupTestServerWith(t *testing.T, mutate func(*config.Config)) (*httptest.Server, *db.DB) {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	xrayPath := filepath.Join(dir, "xray", "config.json")

	cfg := &config.Config{
		Server:   config.ServerConfig{Port: 8080, APIKey: "test-secret-key16"},
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
	if mutate != nil {
		mutate(cfg)
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
		req.Header.Set("X-API-Key", "test-secret-key16")
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
		req.Header.Set("X-API-Key", "test-secret-key16")
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
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("register without auth: expected 401, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	req, _ = http.NewRequest(http.MethodPost, base+"/client/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Version", "v1")
	authHeader(req)
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
		Profiles []struct {
			Profile   string `json:"profile"`
			Transport string `json:"transport"`
			Priority  int    `json:"priority"`
			Tags      string `json:"tags"`
		} `json:"profiles"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&cfg1); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if cfg1.ConfigHash == "" || len(cfg1.Servers) == 0 {
		t.Fatalf("invalid config: %+v", cfg1)
	}
	if len(cfg1.Profiles) == 0 || cfg1.Profiles[0].Priority < 1 || cfg1.Profiles[0].Transport == "" {
		t.Fatalf("invalid profiles in config: %+v", cfg1.Profiles)
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

	req, _ = http.NewRequest(http.MethodGet, base+"/client/commands?timeout=1&stream=sse", nil)
	deviceHeader(req)
	req.Header.Set("Accept", "text/event-stream")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	sseBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("commands sse: expected 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(string(sseBody), "event: commands") {
		t.Fatalf("commands sse: missing event, body=%s", sseBody)
	}
}

func TestInviteTokenRegistration(t *testing.T) {
	if os.Getenv("CGO_ENABLED") == "0" {
		t.Skip("sqlite requires CGO")
	}
	srv, _ := setupTestServer(t)
	defer srv.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	base := srv.URL + "/api"

	body := []byte(`{"email":"invitee@example.com","traffic_gb":5,"expire_days":14}`)
	req, _ := http.NewRequest(http.MethodPost, base+"/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-secret-key16")
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

	invBody := []byte(`{"label":"laptop","max_uses":1,"expires_hours":24}`)
	req, _ = http.NewRequest(http.MethodPost, base+"/users/"+strconv.FormatUint(uint64(user.ID), 10)+"/invites", bytes.NewReader(invBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-secret-key16")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create invite: %d", resp.StatusCode)
	}
	var inv struct {
		Token string `json:"token"`
		ID    uint   `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&inv); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	regBody := []byte(`{"invite_token":"` + inv.Token + `","label":"from-invite"}`)
	req, _ = http.NewRequest(http.MethodPost, base+"/client/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register with invite: expected 201, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Second use must fail (max_uses=1).
	req, _ = http.NewRequest(http.MethodPost, base+"/client/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode == http.StatusCreated {
		t.Fatal("expected invite reuse to fail")
	}
	resp.Body.Close()
}

// TestClientStatsQuotaTrust verifies that client-reported bytes do NOT raise
// used_bytes by default (honor system), and do when limits.trust_client_stats
// is explicitly enabled (issue #47).
func TestClientStatsQuotaTrust(t *testing.T) {
	if os.Getenv("CGO_ENABLED") == "0" {
		t.Skip("sqlite requires CGO")
	}

	reportStats := func(t *testing.T, srv *httptest.Server, database *db.DB) int64 {
		t.Helper()
		client := &http.Client{Timeout: 5 * time.Second}
		base := srv.URL + "/api"

		req, _ := http.NewRequest(http.MethodPost, base+"/users",
			bytes.NewReader([]byte(`{"email":"quota@example.com","traffic_gb":50,"expire_days":30}`)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-secret-key16")
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
		req.Header.Set("X-API-Key", "test-secret-key16")
		resp, err = client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		var reg struct {
			DeviceToken string `json:"device_token"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&reg); err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()

		statsBody := []byte(`{"session_id":"sess-quota","bytes_in":1000,"bytes_out":2000,"status":"connected"}`)
		req, _ = http.NewRequest(http.MethodPost, base+"/client/stats", bytes.NewReader(statsBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Device-Token", reg.DeviceToken)
		resp, err = client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("stats: expected 200, got %d", resp.StatusCode)
		}

		u, err := database.GetUser(user.ID)
		if err != nil {
			t.Fatal(err)
		}
		return u.UsedBytes
	}

	// Default (trust off): self-reported bytes must not raise the quota counter.
	srvOff, dbOff := setupTestServerWith(t, nil)
	if got := reportStats(t, srvOff, dbOff); got != 0 {
		srvOff.Close()
		t.Fatalf("trust off: expected used_bytes 0, got %d", got)
	}
	srvOff.Close()

	// Trust on: legacy behaviour — client reports raise used_bytes.
	srvOn, dbOn := setupTestServerWith(t, func(cfg *config.Config) {
		cfg.Limits.TrustClientStats = true
	})
	if got := reportStats(t, srvOn, dbOn); got != 3000 {
		srvOn.Close()
		t.Fatalf("trust on: expected used_bytes 3000, got %d", got)
	}
	srvOn.Close()
}
