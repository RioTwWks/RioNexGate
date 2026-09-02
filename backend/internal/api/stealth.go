package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"rionexgate/internal/config"
)

type TransportPresetDTO struct { Enabled bool `json:"enabled"`; Port int `json:"port"` }
type FragmentationDTO struct {
	Enabled bool `json:"enabled"`
	Strategy string `json:"strategy,omitempty"`
	Length string `json:"length,omitempty"`
	Delay string `json:"delay,omitempty"`
	MaxSplit string `json:"max_split,omitempty"`
	Applicable bool `json:"applicable"`
	Limitation string `json:"limitation,omitempty"`
	Aggressive bool `json:"aggressive"`
}
type StealthSettingsDTO struct {
	Presets struct {
		XHTTPReality TransportPresetDTO `json:"xhttp_reality"`
		VisionReality TransportPresetDTO `json:"vision_reality"`
		TLS TransportPresetDTO `json:"tls"`
		AmneziaWG TransportPresetDTO `json:"amneziawg"`
	} `json:"presets"`
	Reality struct {
		Dest string `json:"dest"`
		ServerNames []string `json:"server_names"`
		Fingerprint string `json:"fingerprint"`
		ShortIDs []string `json:"short_ids"`
		PrivateKey string `json:"private_key,omitempty"`
	} `json:"reality"`
	Fragmentation FragmentationDTO `json:"fragmentation"`
}
type DestTestRequest struct { Dest string `json:"dest"` }
type DestTestResponse struct {
	Reachable bool `json:"reachable"`
	StatusCode int `json:"status_code,omitempty"`
	LatencyMS int64 `json:"latency_ms,omitempty"`
	Error string `json:"error,omitempty"`
}

func stealthSettingsFromConfig(s *config.StealthConfig) StealthSettingsDTO {
	dto := StealthSettingsDTO{}
	if s == nil { return dto }
	dto.Presets.XHTTPReality = TransportPresetDTO{Enabled: s.XHTTP.Enabled, Port: s.XHTTP.Port}
	dto.Presets.VisionReality = TransportPresetDTO{Enabled: s.Vision.Enabled, Port: s.Vision.Port}
	dto.Presets.TLS = TransportPresetDTO{Enabled: s.TLS.Enabled, Port: s.TLS.Port}
	dto.Presets.AmneziaWG = TransportPresetDTO{Enabled: s.AWG.Enabled, Port: s.AWG.PortOrDefault()}
	dto.Reality.Dest = s.Reality.Dest
	dto.Reality.ServerNames = append([]string(nil), s.Reality.ServerNames...)
	dto.Reality.Fingerprint = s.FingerprintOrDefault()
	dto.Reality.ShortIDs = append([]string(nil), s.Reality.ShortIDs...)
	dto.Reality.PrivateKey = s.Reality.PrivateKey
	f := s.Fragmentation
	dto.Fragmentation.Enabled = f.IsEnabled()
	dto.Fragmentation.Strategy = f.StrategyOrDefault()
	dto.Fragmentation.Length = f.LengthOrDefault()
	dto.Fragmentation.Delay = f.DelayOrDefault()
	dto.Fragmentation.MaxSplit = f.MaxSplitOrDefault()
	dto.Fragmentation.Applicable = s.FragmentationApplicable()
	dto.Fragmentation.Limitation = s.FragmentationLimitation()
	dto.Fragmentation.Aggressive = f.Aggressive()
	return dto
}

func applyStealthSettings(s *config.StealthConfig, dto StealthSettingsDTO) {
	if s == nil {
		return
	}
	s.Enabled = dto.Presets.XHTTPReality.Enabled || dto.Presets.VisionReality.Enabled || dto.Presets.TLS.Enabled || dto.Presets.AmneziaWG.Enabled
	s.XHTTP.Enabled = dto.Presets.XHTTPReality.Enabled
	if dto.Presets.XHTTPReality.Port > 0 { s.XHTTP.Port = dto.Presets.XHTTPReality.Port }
	s.Vision.Enabled = dto.Presets.VisionReality.Enabled
	if dto.Presets.VisionReality.Port > 0 { s.Vision.Port = dto.Presets.VisionReality.Port }
	s.TLS.Enabled = dto.Presets.TLS.Enabled
	if dto.Presets.TLS.Port > 0 { s.TLS.Port = dto.Presets.TLS.Port }
	s.AWG.Enabled = dto.Presets.AmneziaWG.Enabled
	if dto.Presets.AmneziaWG.Port > 0 { s.AWG.Port = dto.Presets.AmneziaWG.Port }
	s.Reality.Dest = strings.TrimSpace(dto.Reality.Dest)
	s.Reality.ServerNames = append([]string(nil), dto.Reality.ServerNames...)
	s.Reality.ShortIDs = append([]string(nil), dto.Reality.ShortIDs...)
	if dto.Reality.Fingerprint != "" { s.Fingerprint = dto.Reality.Fingerprint }
	if dto.Reality.PrivateKey != "" { s.Reality.PrivateKey = dto.Reality.PrivateKey }
	s.Fragmentation.Enabled = dto.Fragmentation.Enabled
	if dto.Fragmentation.Strategy != "" { s.Fragmentation.Strategy = dto.Fragmentation.Strategy }
	if dto.Fragmentation.Length != "" { s.Fragmentation.Length = dto.Fragmentation.Length }
	if dto.Fragmentation.Delay != "" { s.Fragmentation.Delay = dto.Fragmentation.Delay }
	if dto.Fragmentation.MaxSplit != "" { s.Fragmentation.MaxSplit = dto.Fragmentation.MaxSplit }
}

func (h *Handler) GetStealthSettings(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, stealthSettingsFromConfig(&h.cfg.Core.Stealth))
}
func (h *Handler) UpdateStealthSettings(w http.ResponseWriter, r *http.Request) {
	var dto StealthSettingsDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json"); return
	}
	applyStealthSettings(&h.cfg.Core.Stealth, dto)
	if err := h.core.Reload(); err != nil {
		writeError(w, http.StatusInternalServerError, "settings saved but core reload failed: "+err.Error()); return
	}
	writeJSON(w, http.StatusOK, stealthSettingsFromConfig(&h.cfg.Core.Stealth))
}
func (h *Handler) TestStealthDest(w http.ResponseWriter, r *http.Request) {
	var req DestTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json"); return
	}
	dest := strings.TrimSpace(req.Dest)
	if dest == "" { writeError(w, http.StatusBadRequest, "dest is required"); return }
	if !strings.Contains(dest, ":") { dest += ":443" }
	host := dest
	if idx := strings.Index(dest, ":"); idx > 0 { host = dest[:idx] }
	start := time.Now()
	client := &http.Client{Timeout: 10 * time.Second}
	url := "https://" + host
	reqHTTP, err := http.NewRequestWithContext(r.Context(), http.MethodHead, url, nil)
	if err != nil { writeJSON(w, http.StatusOK, DestTestResponse{Reachable: false, Error: err.Error()}); return }
	resp, err := client.Do(reqHTTP)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		reqHTTP, _ = http.NewRequestWithContext(r.Context(), http.MethodGet, url, nil)
		resp, err = client.Do(reqHTTP)
		latency = time.Since(start).Milliseconds()
		if err != nil { writeJSON(w, http.StatusOK, DestTestResponse{Reachable: false, Error: err.Error(), LatencyMS: latency}); return }
	}
	defer resp.Body.Close()
	writeJSON(w, http.StatusOK, DestTestResponse{Reachable: true, StatusCode: resp.StatusCode, LatencyMS: latency})
}
