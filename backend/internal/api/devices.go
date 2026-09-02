package api

import (
	"net/http"
	"time"

	"rionexgate/internal/models"

	"github.com/go-chi/chi/v5"
)

type DeviceDTO struct {
	ID         uint       `json:"id"`
	UserID     uint       `json:"user_id"`
	Token      string     `json:"token"`
	Label      string     `json:"label"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

func toDeviceDTO(d models.Device) DeviceDTO {
	return DeviceDTO{
		ID:         d.ID,
		UserID:     d.UserID,
		Token:      d.Token,
		Label:      d.Label,
		LastSeenAt: d.LastSeenAt,
		CreatedAt:  d.CreatedAt,
	}
}

func (h *Handler) ListUserDevices(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserParam(r)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	devices, err := h.db.ListDevicesByUser(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	dtos := make([]DeviceDTO, len(devices))
	for i, d := range devices {
		dtos[i] = toDeviceDTO(d)
	}
	writeJSON(w, http.StatusOK, dtos)
}

func (h *Handler) RevokeUserDevice(w http.ResponseWriter, r *http.Request) {
	userID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	deviceID, err := parseID(chi.URLParam(r, "deviceId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid device id")
		return
	}
	device, err := h.db.GetDeviceByIDForUser(userID, deviceID)
	if err != nil {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	h.commands.Enqueue(device.Token, ClientCommand{Type: "disconnect"})
	if err := h.db.DeleteDevice(device.Token); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
