package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"rionexgate/internal/config"
)

// TestStealthDestRespectsPort ensures the dest port is actually dialed:
// a TLS listener on a custom port must produce a TLS/certificate error,
// while the old implementation always dialed :443 (connection refused).
func TestStealthDestRespectsPort(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	dest := strings.TrimPrefix(srv.URL, "https://") // 127.0.0.1:<port>

	h := &Handler{cfg: &config.Config{}}
	h.cfg.Server.AllowPrivateProbes = true

	req := httptest.NewRequest(http.MethodPost, "/api/stealth/test-dest", strings.NewReader(`{"dest":"`+dest+`"}`))
	rec := httptest.NewRecorder()
	h.TestStealthDest(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var resp DestTestResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	// Self-signed cert => handshake completes but verification fails.
	// The point: error must come from the TLS dial to the custom port,
	// not "connection refused" on :443.
	if resp.Reachable {
		t.Fatal("self-signed test server must not be reported as reachable")
	}
	if !strings.Contains(resp.Error, "certificate") {
		t.Fatalf("expected TLS certificate error from custom port %s, got: %s", dest, resp.Error)
	}
}
