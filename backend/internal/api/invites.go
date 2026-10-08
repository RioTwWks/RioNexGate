package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"rionexgate/internal/db"
	"rionexgate/internal/models"

	"github.com/go-chi/chi/v5"
)

type InviteDTO struct {
	ID        uint       `json:"id"`
	UserID    uint       `json:"user_id"`
	Token     string     `json:"token"`
	Label     string     `json:"label,omitempty"`
	MaxUses   int        `json:"max_uses"`
	UsedCount int        `json:"used_count"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	Usable    bool       `json:"usable"`
	CreatedAt time.Time  `json:"created_at"`
}

func toInviteDTO(inv models.Invite) InviteDTO {
	return InviteDTO{
		ID:        inv.ID,
		UserID:    inv.UserID,
		Token:     inv.Token,
		Label:     inv.Label,
		MaxUses:   inv.MaxUses,
		UsedCount: inv.UsedCount,
		ExpiresAt: inv.ExpiresAt,
		RevokedAt: inv.RevokedAt,
		Usable:    inv.Usable(),
		CreatedAt: inv.CreatedAt,
	}
}

type CreateInviteRequest struct {
	Label        string `json:"label"`
	MaxUses      int    `json:"max_uses"`
	ExpiresHours int    `json:"expires_hours"`
}

func (h *Handler) ListUserInvites(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserParam(r)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	invites, err := h.db.ListInvitesByUser(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]InviteDTO, len(invites))
	for i, inv := range invites {
		out[i] = toInviteDTO(inv)
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) CreateUserInvite(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserParam(r)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	var req CreateInviteRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxJSONBody)).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	in := db.CreateInviteInput{
		UserID:  user.ID,
		Label:   req.Label,
		MaxUses: req.MaxUses,
	}
	if req.ExpiresHours > 0 {
		in.ExpiresIn = time.Duration(req.ExpiresHours) * time.Hour
	}
	inv, err := h.db.CreateInvite(in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, toInviteDTO(*inv))
}

func (h *Handler) RevokeUserInvite(w http.ResponseWriter, r *http.Request) {
	userID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	inviteID, err := parseID(chi.URLParam(r, "inviteId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid invite id")
		return
	}
	if err := h.db.RevokeInvite(userID, inviteID); err != nil {
		writeError(w, http.StatusNotFound, "invite not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
