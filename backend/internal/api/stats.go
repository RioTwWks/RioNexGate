package api

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
)

type StatsPoint struct {
	Time      time.Time `json:"time"`
	BytesUp   int64     `json:"bytes_up"`
	BytesDown int64     `json:"bytes_down"`
}

type TotalStatsResponse struct {
	TotalUsedGB float64      `json:"total_used_gb"`
	Points      []StatsPoint `json:"points"`
}

type ClientMetricsResponse struct {
	ActiveDevices     int64 `json:"active_devices"`
	SyncRequests      int64 `json:"sync_requests"`
	RegistrationFails int64 `json:"registration_fails"`
}

func (h *Handler) GetClientMetrics(w http.ResponseWriter, r *http.Request) {
	since := time.Now().Add(-24 * time.Hour)
	active, _ := h.db.CountActiveDevices(since)
	writeJSON(w, http.StatusOK, ClientMetricsResponse{
		ActiveDevices:     active,
		SyncRequests:      atomic.LoadInt64(&h.metrics.SyncRequests),
		RegistrationFails: atomic.LoadInt64(&h.metrics.RegistrationFails),
	})
}

func (h *Handler) GetTotalStats(w http.ResponseWriter, r *http.Request) {
	total, err := h.db.TotalUsedBytes()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	since := time.Now().AddDate(0, 0, -7)
	records, err := h.db.TrafficHistory(since)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	points := make([]StatsPoint, len(records))
	for i, rec := range records {
		points[i] = StatsPoint{
			Time:      rec.RecordedAt,
			BytesUp:   rec.BytesUp,
			BytesDown: rec.BytesDown,
		}
	}
	writeJSON(w, http.StatusOK, TotalStatsResponse{
		TotalUsedGB: float64(total) / (1024 * 1024 * 1024),
		Points:      points,
	})
}

func (h *Handler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	user, err := h.db.GetUser(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	usedGB, err := h.core.GetStats(user.UUID)
	if err != nil {
		usedGB = float64(user.UsedBytes) / (1024 * 1024 * 1024)
	}

	clientBytes, _ := h.db.ClientReportedBytesForUser(id)
	clientGB := float64(clientBytes) / (1024 * 1024 * 1024)
	serverBytes := int64(usedGB * 1024 * 1024 * 1024)
	if clientBytes > serverBytes {
		usedGB = clientGB
	}

	since := time.Now().AddDate(0, 0, -7)
	records, _ := h.db.UserTrafficHistory(id, since)
	points := make([]StatsPoint, len(records))
	for i, rec := range records {
		points[i] = StatsPoint{
			Time:      rec.RecordedAt,
			BytesUp:   rec.BytesUp,
			BytesDown: rec.BytesDown,
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":            user.ID,
		"used_gb":            usedGB,
		"limit_gb":           user.TrafficGB,
		"client_reported_gb": clientGB,
		"server_reported_gb": float64(serverBytes) / (1024 * 1024 * 1024),
		"points":             points,
	})
}
