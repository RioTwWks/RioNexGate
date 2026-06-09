package api

import (
	"net/http"
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
		"user_id":  user.ID,
		"used_gb":  usedGB,
		"limit_gb": user.TrafficGB,
		"points":   points,
	})
}
