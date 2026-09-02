package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"rionexgate/internal/config"
	"rionexgate/internal/db"
	"rionexgate/internal/models"
)

type Manager interface {
	Reload() error
	GetStats(userID string) (float64, error)
	GetClientLink(userID string, protocol string) (string, error)
	Type() string
	SetType(t string) error
	StartStatsCollector(ctx context.Context)
}

type manager struct {
	cfg    *config.Config
	db     *db.DB
	mu     sync.RWMutex
}

func NewManager(cfg *config.Config, database *db.DB) Manager {
	return &manager{cfg: cfg, db: database}
}

func (m *manager) Type() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg.Core.Type
}

func (m *manager) SetType(t string) error {
	if t != "xray" && t != "sing-box" {
		return fmt.Errorf("unsupported core type: %s", t)
	}
	m.mu.Lock()
	m.cfg.SetCoreType(t)
	m.mu.Unlock()
	return m.Reload()
}

func (m *manager) configPath() string {
	if m.Type() == "sing-box" {
		return m.cfg.Core.Singbox.ConfigPath
	}
	return m.cfg.Core.Xray.ConfigPath
}

func (m *manager) Reload() error {
	users, err := m.db.ListActiveUsers()
	if err != nil {
		return err
	}

	var data []byte
	listenPort := m.cfg.Core.ListenPort
	if m.Type() == "sing-box" {
		data, err = generateSingboxConfig(listenPort, m.cfg.Core.Singbox.APIAddress, users)
	} else {
		data, err = generateXrayConfig(listenPort, m.cfg.Core.Xray.APIAddress, users)
	}
	if err != nil {
		return err
	}

	path := m.configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	log.Printf("core config written to %s (%d users)", path, len(users))
	return nil
}

func (m *manager) GetStats(userID string) (float64, error) {
	user, err := m.db.GetUserByUUID(userID)
	if err != nil {
		// try numeric id
		var id uint
		if _, scanErr := fmt.Sscanf(userID, "%d", &id); scanErr == nil {
			user, err = m.db.GetUser(id)
		}
	}
	if err != nil {
		return 0, err
	}
	return float64(user.UsedBytes) / (1024 * 1024 * 1024), nil
}

func (m *manager) GetClientLink(userID string, protocol string) (string, error) {
	var user *models.User
	var err error
	var id uint
	if _, scanErr := fmt.Sscanf(userID, "%d", &id); scanErr == nil {
		user, err = m.db.GetUser(id)
	} else {
		user, err = m.db.GetUserByUUID(userID)
	}
	if err != nil {
		return "", err
	}
	link := GetClientLink(m.cfg.Core.PublicHost, m.cfg.Core.ListenPort, *user, protocol)
	return link, nil
}

func (m *manager) StartStatsCollector(ctx context.Context) {
	interval := time.Duration(m.cfg.Core.StatsPoll) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.collectStats()
		}
	}
}

func (m *manager) collectStats() {
	users, err := m.db.ListUsers()
	if err != nil {
		log.Printf("stats: list users: %v", err)
		return
	}
	for _, user := range users {
		up, down, err := m.fetchUserStats(user.Email)
		if err != nil {
			continue
		}
		total := up + down
		if total == 0 {
			continue
		}
		if err := m.db.UpdateUserUsedBytes(user.ID, total); err != nil {
			log.Printf("stats: update user %d: %v", user.ID, err)
			continue
		}
		_ = m.db.RecordTraffic(user.ID, up, down)
	}
}

func (m *manager) fetchUserStats(email string) (int64, int64, error) {
	if m.Type() != "xray" {
		return 0, 0, fmt.Errorf("stats not implemented for %s", m.Type())
	}
	apiAddr := m.cfg.Core.Xray.APIAddress
	if apiAddr == "" {
		return 0, 0, fmt.Errorf("xray api address not configured")
	}

	up, err := queryXrayStat(apiAddr, "user>>>"+email+">>>traffic>>>uplink")
	if err != nil {
		return 0, 0, err
	}
	down, err := queryXrayStat(apiAddr, "user>>>"+email+">>>traffic>>>downlink")
	if err != nil {
		return 0, 0, err
	}
	return up, down, nil
}

type xrayStatsResponse struct {
	Stat map[string]int64 `json:"stat"`
}

func queryXrayStat(apiAddr, name string) (int64, error) {
	url := fmt.Sprintf("http://%s/stats?name=%s", apiAddr, name)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("xray stats: %s", string(body))
	}
	var result struct {
		Stat []struct {
			Name  string `json:"name"`
			Value int64  `json:"value"`
		} `json:"stat"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		// try alternate format
		var alt struct {
			Value int64 `json:"value"`
		}
		if err2 := json.Unmarshal(body, &alt); err2 != nil {
			return 0, err
		}
		return alt.Value, nil
	}
	for _, s := range result.Stat {
		if s.Name == name {
			return s.Value, nil
		}
	}
	if len(result.Stat) > 0 {
		return result.Stat[0].Value, nil
	}
	return 0, nil
}
