package api

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"rionexgate/internal/config"
)

type RotateAPIKeyRequest struct {
	NewKey string `json:"new_key"`
}

type RotateAPIKeyResponse struct {
	APIKey string `json:"api_key"`
}

type MintSubscriptionLinkRequest struct {
	TTLHours int `json:"ttl_hours"`
}

type MintSubscriptionLinkResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
	TTLHours  int       `json:"ttl_hours"`
}

func (h *Handler) RotateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req RotateAPIKeyRequest
	_ = json.NewDecoder(http.MaxBytesReader(w, r.Body, maxJSONBody)).Decode(&req)

	newKey := strings.TrimSpace(req.NewKey)
	if newKey == "" {
		generated, err := config.GenerateAPIKey()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		newKey = generated
	}
	if err := h.cfg.RotateAPIKey(newKey); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, RotateAPIKeyResponse{APIKey: newKey})
}

func (h *Handler) MintUserSubscriptionLink(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserParam(r)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	var req MintSubscriptionLinkRequest
	_ = json.NewDecoder(http.MaxBytesReader(w, r.Body, maxJSONBody)).Decode(&req)

	token, err := h.db.EnsureSubscriptionToken(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ttlHours := req.TTLHours
	if ttlHours <= 0 {
		ttlHours = h.cfg.Server.SubscriptionDefaultTTLHours
	}
	if ttlHours > 24*365 {
		ttlHours = 24 * 365
	}
	base, err := h.publicBase(r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	url, exp, err := h.signedSubscriptionURL(base, token, time.Duration(ttlHours)*time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, MintSubscriptionLinkResponse{
		URL:       url,
		ExpiresAt: exp.UTC(),
		TTLHours:  ttlHours,
	})
}

// errHostNotAllowed is returned when a request Host is outside the configured
// server.allowed_hosts whitelist and no server.public_base_url is set, so no
// trustworthy base URL can be built for generated links.
var errHostNotAllowed = errors.New("request host is not in server.allowed_hosts and server.public_base_url is not set")

// publicBase resolves the absolute base URL for generated links.
//
// Priority: server.public_base_url (recommended for production). Without it the
// base is derived from request headers (Host, X-Forwarded-Proto) — accepted only
// when the Host matches server.allowed_hosts if that whitelist is configured;
// otherwise the legacy behaviour is kept with a one-time startup-style warning.
func (h *Handler) publicBase(r *http.Request) (string, error) {
	base := strings.TrimSpace(h.cfg.Server.PublicBaseURL)
	if base != "" {
		return strings.TrimRight(base, "/"), nil
	}
	if len(h.cfg.Server.AllowedHosts) > 0 && !hostAllowed(r.Host, h.cfg.Server.AllowedHosts) {
		log.Printf("public base: refusing header-derived URL for host %q (not in server.allowed_hosts); set server.public_base_url", r.Host)
		return "", errHostNotAllowed
	}
	h.publicBaseWarnOnce.Do(func() {
		log.Printf("warning: server.public_base_url is not set; absolute links (subscription URLs) are built from request " +
			"Host/X-Forwarded-Proto headers, which is unsafe behind untrusted proxies — set public_base_url " +
			"(and optionally server.allowed_hosts) for production")
	})
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if fwd := r.Header.Get("X-Forwarded-Proto"); fwd != "" {
		scheme = fwd
	}
	return scheme + "://" + r.Host, nil
}

// hostAllowed reports whether host matches the whitelist, comparing both the
// full host (with port) and the bare hostname.
func hostAllowed(host string, allowed []string) bool {
	if host == "" {
		return false
	}
	name := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		name = h
	}
	for _, a := range allowed {
		a = strings.TrimSpace(a)
		if a == "" {
			continue
		}
		if strings.EqualFold(a, host) || strings.EqualFold(a, name) {
			return true
		}
	}
	return false
}
