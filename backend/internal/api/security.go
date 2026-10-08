package api

import (
	"encoding/json"
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
	base := h.publicBase(r)
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

func (h *Handler) publicBase(r *http.Request) string {
	base := h.cfg.Server.PublicBaseURL
	if base != "" {
		return strings.TrimRight(base, "/")
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if fwd := r.Header.Get("X-Forwarded-Proto"); fwd != "" {
		scheme = fwd
	}
	return scheme + "://" + r.Host
}
