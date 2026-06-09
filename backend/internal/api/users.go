package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"proxy-mgr/internal/db"
	"proxy-mgr/internal/models"

	"github.com/go-chi/chi/v5"
)

type UserDTO struct {
	ID        uint      `json:"id"`
	UUID      string    `json:"uuid"`
	Email     string    `json:"email"`
	TrafficGB int64     `json:"traffic_gb"`
	UsedGB    float64   `json:"used_gb"`
	ExpiresAt time.Time `json:"expires_at"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

func toUserDTO(u models.User) UserDTO {
	return UserDTO{
		ID:        u.ID,
		UUID:      u.UUID,
		Email:     u.Email,
		TrafficGB: u.TrafficGB,
		UsedGB:    float64(u.UsedBytes) / (1024 * 1024 * 1024),
		ExpiresAt: u.ExpiresAt,
		Active:    u.Active,
		CreatedAt: u.CreatedAt,
	}
}

type CreateUserRequest struct {
	Email      string `json:"email"`
	TrafficGB  int64  `json:"traffic_gb"`
	ExpireDays int    `json:"expire_days"`
}

type UpdateUserRequest struct {
	Email      *string    `json:"email"`
	TrafficGB  *int64     `json:"traffic_gb"`
	ExpireDays *int       `json:"expire_days"`
	Active     *bool      `json:"active"`
	ExpiresAt  *time.Time `json:"expires_at"`
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.db.ListUsers()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	dtos := make([]UserDTO, len(users))
	for i, u := range users {
		dtos[i] = toUserDTO(u)
	}
	writeJSON(w, http.StatusOK, dtos)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.TrafficGB == 0 {
		req.TrafficGB = h.cfg.Limits.DefaultTrafficGB
	}
	if req.ExpireDays == 0 {
		req.ExpireDays = h.cfg.Limits.DefaultExpireDays
	}
	user, err := h.db.CreateUser(db.CreateUserInput{
		Email:      req.Email,
		TrafficGB:  req.TrafficGB,
		ExpireDays: req.ExpireDays,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.core.Reload(); err != nil {
		writeError(w, http.StatusInternalServerError, "user created but reload failed: "+err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, toUserDTO(*user))
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserParam(r)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, toUserDTO(*user))
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	in := db.UpdateUserInput{
		Email:     req.Email,
		TrafficGB: req.TrafficGB,
		Active:    req.Active,
		ExpiresAt: req.ExpiresAt,
	}
	if req.ExpireDays != nil {
		t := time.Now().AddDate(0, 0, *req.ExpireDays)
		in.ExpiresAt = &t
	}
	user, err := h.db.UpdateUser(id, in)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err := h.core.Reload(); err != nil {
		writeError(w, http.StatusInternalServerError, "updated but reload failed: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toUserDTO(*user))
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.db.DeleteUser(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err := h.core.Reload(); err != nil {
		writeError(w, http.StatusInternalServerError, "deleted but reload failed: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetUserLink(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserParam(r)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	proto := r.URL.Query().Get("proto")
	if proto == "" {
		proto = "vless"
	}
	link, err := h.core.GetClientLink(strconv.FormatUint(uint64(user.ID), 10), proto)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"link": link, "protocol": proto})
}

func (h *Handler) GetUserQR(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserParam(r)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	proto := r.URL.Query().Get("proto")
	if proto == "" {
		proto = "vless"
	}
	link, err := h.core.GetClientLink(strconv.FormatUint(uint64(user.ID), 10), proto)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	png, err := generateQR(link)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(png)
}

func (h *Handler) getUserParam(r *http.Request) (*models.User, error) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		return nil, err
	}
	return h.db.GetUser(id)
}

func parseID(s string) (uint, error) {
	id, err := strconv.ParseUint(s, 10, 64)
	return uint(id), err
}
