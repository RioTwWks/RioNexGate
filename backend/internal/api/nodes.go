package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"rionexgate/internal/db"
	"rionexgate/internal/models"

	"github.com/go-chi/chi/v5"
)

type NodeDTO struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	Port        int    `json:"port"`
	Active      bool   `json:"active"`
	Role        string `json:"role"`
	Protocol    string `json:"protocol"`
	Credentials string `json:"credentials,omitempty"`
	Region      string `json:"region,omitempty"`
	Priority    int    `json:"priority"`
}

func toNodeDTO(n models.Node) NodeDTO {
	return NodeDTO{
		ID:          n.ID,
		Name:        n.Name,
		Address:     n.Address,
		Port:        n.Port,
		Active:      n.Active,
		Role:        n.Role,
		Protocol:    n.Protocol,
		Credentials: n.Credentials,
		Region:      n.Region,
		Priority:    n.Priority,
	}
}

type CreateNodeRequest struct {
	Name        string `json:"name"`
	Address     string `json:"address"`
	Port        int    `json:"port"`
	Active      *bool  `json:"active"`
	Role        string `json:"role"`
	Protocol    string `json:"protocol"`
	Credentials string `json:"credentials"`
	Region      string `json:"region"`
	Priority    int    `json:"priority"`
}

type UpdateNodeRequest struct {
	Name        *string `json:"name"`
	Address     *string `json:"address"`
	Port        *int    `json:"port"`
	Active      *bool   `json:"active"`
	Role        *string `json:"role"`
	Protocol    *string `json:"protocol"`
	Credentials *string `json:"credentials"`
	Region      *string `json:"region"`
	Priority    *int    `json:"priority"`
}

type UserChainRequest struct {
	EntryNodeID *uint `json:"entry_node_id"`
	ExitNodeID  *uint `json:"exit_node_id"`
	Clear       bool  `json:"clear"`
}

type NodeHealthResponse struct {
	Reachable bool   `json:"reachable"`
	CheckType string `json:"check_type"`
	LatencyMS int64  `json:"latency_ms,omitempty"`
	Error     string `json:"error,omitempty"`
}

func (h *Handler) ListNodes(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.db.ListNodes()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	dtos := make([]NodeDTO, len(nodes))
	for i, n := range nodes {
		dtos[i] = toNodeDTO(n)
	}
	writeJSON(w, http.StatusOK, dtos)
}

func (h *Handler) CreateNode(w http.ResponseWriter, r *http.Request) {
	var req CreateNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}
	node, err := h.db.CreateNode(db.CreateNodeInput{
		Name:        req.Name,
		Address:     req.Address,
		Port:        req.Port,
		Active:      active,
		Role:        req.Role,
		Protocol:    req.Protocol,
		Credentials: req.Credentials,
		Region:      req.Region,
		Priority:    req.Priority,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.core.Reload(); err != nil {
		writeError(w, http.StatusInternalServerError, "node created but reload failed: "+err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, toNodeDTO(*node))
}

func (h *Handler) GetNode(w http.ResponseWriter, r *http.Request) {
	node, err := h.getNodeParam(r)
	if err != nil {
		writeError(w, http.StatusNotFound, "node not found")
		return
	}
	writeJSON(w, http.StatusOK, toNodeDTO(*node))
}

func (h *Handler) UpdateNode(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req UpdateNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	node, err := h.db.UpdateNode(id, db.UpdateNodeInput{
		Name:        req.Name,
		Address:     req.Address,
		Port:        req.Port,
		Active:      req.Active,
		Role:        req.Role,
		Protocol:    req.Protocol,
		Credentials: req.Credentials,
		Region:      req.Region,
		Priority:    req.Priority,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err := h.core.Reload(); err != nil {
		writeError(w, http.StatusInternalServerError, "updated but reload failed: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toNodeDTO(*node))
}

func (h *Handler) CheckNodeHealth(w http.ResponseWriter, r *http.Request) {
	node, err := h.getNodeParam(r)
	if err != nil {
		writeError(w, http.StatusNotFound, "node not found")
		return
	}
	writeJSON(w, http.StatusOK, checkNodeTCP(r.Context(), node.Address, node.Port))
}

func checkNodeTCP(ctx context.Context, address string, port int) NodeHealthResponse {
	if address == "" {
		return NodeHealthResponse{CheckType: "tcp", Error: "address is empty"}
	}
	if port <= 0 || port > 65535 {
		return NodeHealthResponse{CheckType: "tcp", Error: "invalid port"}
	}
	target := net.JoinHostPort(address, fmt.Sprintf("%d", port))
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	start := time.Now()
	conn, err := dialer.DialContext(ctx, "tcp", target)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return NodeHealthResponse{CheckType: "tcp", Error: err.Error(), LatencyMS: latency}
	}
	_ = conn.Close()
	return NodeHealthResponse{Reachable: true, CheckType: "tcp", LatencyMS: latency}
}

func (h *Handler) DeleteNode(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.db.DeleteNode(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err := h.core.Reload(); err != nil {
		writeError(w, http.StatusInternalServerError, "deleted but reload failed: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpdateUserChain(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req UserChainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.EntryNodeID != nil {
		node, err := h.db.GetNode(*req.EntryNodeID)
		if err != nil || node.Role != models.NodeRoleEntry {
			writeError(w, http.StatusBadRequest, "invalid entry node")
			return
		}
	}
	if req.ExitNodeID != nil {
		node, err := h.db.GetNode(*req.ExitNodeID)
		if err != nil || node.Role != models.NodeRoleExit {
			writeError(w, http.StatusBadRequest, "invalid exit node")
			return
		}
	}
	user, err := h.db.UpdateUser(id, db.UpdateUserInput{
		EntryNodeID: req.EntryNodeID,
		ExitNodeID:  req.ExitNodeID,
		ClearChain:  req.Clear,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	h.notifyUserDevicesRefresh(user.ID)
	if err := h.core.Reload(); err != nil {
		writeError(w, http.StatusInternalServerError, "updated but reload failed: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toUserDTO(*user))
}

func (h *Handler) getNodeParam(r *http.Request) (*models.Node, error) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		return nil, err
	}
	return h.db.GetNode(id)
}
