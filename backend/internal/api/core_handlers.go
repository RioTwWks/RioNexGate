package api

import (
	"encoding/json"
	"net/http"
)

type CoreTypeRequest struct {
	Type string `json:"type"`
}

func (h *Handler) ReloadCore(w http.ResponseWriter, r *http.Request) {
	if err := h.core.Reload(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "reloaded"})
}

func (h *Handler) GetCoreType(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"type": h.core.Type()})
}

func (h *Handler) SetCoreType(w http.ResponseWriter, r *http.Request) {
	var req CoreTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.core.SetType(req.Type); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"type": h.core.Type()})
}
