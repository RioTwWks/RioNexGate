package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"rionexgate/internal/api/middleware"
	"rionexgate/internal/core"
	"rionexgate/internal/db"
	"rionexgate/internal/models"

	"github.com/go-chi/chi/v5"
)

type ClientMetrics struct {
	ActiveDevices     int64
	RegistrationFails int64
	SyncRequests      int64
}

type RegisterRequest struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Label  string `json:"label"`
}

type RegisterResponse struct {
	DeviceToken     string `json:"device_token"`
	SubscriptionURL string `json:"subscription_url"`
}

type StatsRequest struct {
	SessionID string `json:"session_id"`
	BytesIn   int64  `json:"bytes_in"`
	BytesOut  int64  `json:"bytes_out"`
	Sessions  int    `json:"sessions"`
	Status    string `json:"status"`
}

func (h *Handler) subscriptionURL(r *http.Request, token string) string {
	base := h.cfg.Server.PublicBaseURL
	if base == "" {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		if fwd := r.Header.Get("X-Forwarded-Proto"); fwd != "" {
			scheme = fwd
		}
		base = scheme + "://" + r.Host
	}
	return base + "/api/subscription/" + token
}

func (h *Handler) RegisterClient(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		atomic.AddInt64(&h.metrics.RegistrationFails, 1)
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	var userID uint
	switch {
	case req.UserID > 0:
		userID = req.UserID
	case req.Email != "":
		user, err := h.db.GetUserByEmail(req.Email)
		if err != nil {
			atomic.AddInt64(&h.metrics.RegistrationFails, 1)
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		userID = user.ID
	default:
		atomic.AddInt64(&h.metrics.RegistrationFails, 1)
		writeError(w, http.StatusBadRequest, "user_id or email is required")
		return
	}

	user, err := h.db.GetUser(userID)
	if err != nil {
		atomic.AddInt64(&h.metrics.RegistrationFails, 1)
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	if !user.Active {
		atomic.AddInt64(&h.metrics.RegistrationFails, 1)
		writeError(w, http.StatusForbidden, "user inactive")
		return
	}

	subToken, err := h.db.EnsureSubscriptionToken(user.ID)
	if err != nil {
		atomic.AddInt64(&h.metrics.RegistrationFails, 1)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	device, err := h.db.CreateDevice(user.ID, req.Label)
	if err != nil {
		atomic.AddInt64(&h.metrics.RegistrationFails, 1)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, RegisterResponse{
		DeviceToken:     device.Token,
		SubscriptionURL: h.subscriptionURL(r, subToken),
	})
}

func (h *Handler) GetClientConfig(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt64(&h.metrics.SyncRequests, 1)
	dc, ok := middleware.DeviceFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	_ = h.db.UpdateDeviceLastSeen(dc.Device.Token)

	cfg, err := h.buildClientConfig(dc.User)
	if err != nil {
		log.Printf("client config generation failed for user %d: %v", dc.User.ID, err)
		if dc.Device.CachedConfig != "" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Config-Cached", "true")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(dc.Device.CachedConfig))
			return
		}
		writeError(w, http.StatusInternalServerError, "config generation failed")
		return
	}

	raw, err := cfg.JSON()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.db.UpdateDeviceConfigCache(dc.Device.ID, string(raw), cfg.ConfigHash); err != nil {
		log.Printf("client config cache update failed for device %d: %v", dc.Device.ID, err)
	}

	writeJSONBytes(w, http.StatusOK, raw)
}

func (h *Handler) buildClientConfig(user *models.User) (*core.ClientConfig, error) {
	entry, _ := h.db.ResolveUserEntryNode(user)
	return core.BuildClientConfig(
		h.cfg.Core.PublicHost,
		h.cfg.Core.ListenPort,
		*user,
		h.cfg.Server.ClientSOCKS5Port,
		&h.cfg.Core.Stealth,
		entry,
	)
}

func (h *Handler) PostClientStats(w http.ResponseWriter, r *http.Request) {
	dc, ok := middleware.DeviceFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req StatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.SessionID == "" {
		writeError(w, http.StatusBadRequest, "session_id is required")
		return
	}

	if err := h.db.UpsertClientStats(db.ClientStatsInput{
		DeviceToken: dc.Device.Token,
		SessionID:   req.SessionID,
		BytesIn:     req.BytesIn,
		BytesOut:    req.BytesOut,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = h.db.UpdateDeviceLastSeen(dc.Device.Token)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) GetClientCommands(w http.ResponseWriter, r *http.Request) {
	dc, ok := middleware.DeviceFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	timeout := 30 * time.Second
	if v := r.URL.Query().Get("timeout"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil && secs > 0 && secs <= 60 {
			timeout = time.Duration(secs) * time.Second
		}
	}

	_ = h.db.UpdateDeviceLastSeen(dc.Device.Token)

	if wantsSSE(r) {
		h.getClientCommandsSSE(w, r, dc.Device.Token, timeout)
		return
	}

	cmds := h.commands.Poll(dc.Device.Token, timeout)
	writeJSON(w, http.StatusOK, map[string]interface{}{"commands": cmds})
}

func wantsSSE(r *http.Request) bool {
	if strings.EqualFold(r.URL.Query().Get("stream"), "sse") {
		return true
	}
	return strings.Contains(r.Header.Get("Accept"), "text/event-stream")
}

func (h *Handler) getClientCommandsSSE(w http.ResponseWriter, r *http.Request, deviceToken string, timeout time.Duration) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	cmds := h.commands.Poll(deviceToken, timeout)
	payload, err := json.Marshal(map[string]interface{}{"commands": cmds})
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"error\":\"marshal failed\"}\n\n")
		flusher.Flush()
		return
	}
	fmt.Fprintf(w, "event: commands\ndata: %s\n\n", payload)
	flusher.Flush()
}

func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "token is required")
		return
	}

	user, err := h.db.GetUserBySubscriptionToken(token)
	if err != nil {
		writeError(w, http.StatusNotFound, "subscription not found")
		return
	}
	if !user.Active {
		writeError(w, http.StatusForbidden, "subscription inactive")
		return
	}

	entry, _ := h.db.ResolveUserEntryNode(user)
	var peer *models.WireGuardPeer
	if h.cfg.Core.Stealth.AWGActive() { peer, _ = h.db.EnsureWireGuardPeer(user.ID, h.cfg.Core.Stealth.AWG.SubnetOrDefault()) }
	payload := core.BuildSubscriptionBase64Graceful(
		h.cfg.Core.PublicHost,
		h.cfg.Core.ListenPort,
		*user,
		&h.cfg.Core.Stealth,
		entry,
		peer,
	)

	if _, err := base64.StdEncoding.DecodeString(payload); err != nil {
		writeError(w, http.StatusInternalServerError, "subscription generation failed")
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(payload))
}

func (h *Handler) notifyUserDevicesRefresh(userID uint) {
	devices, err := h.db.ListDevicesByUser(userID)
	if err != nil {
		return
	}
	for _, dev := range devices {
		h.commands.Enqueue(dev.Token, ClientCommand{Type: "refresh_config"})
	}
}

func writeJSONBytes(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}
