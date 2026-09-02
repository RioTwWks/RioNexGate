package core

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
	"text/template"
	"rionexgate/internal/config"
	"rionexgate/internal/models"
)

type AWGPeerEntry struct {
	UserID uint; Email, PublicKey, PresharedKey, ClientIP string
}
type awgTemplateData struct {
	ServerPrivateKey, ServerAddress string
	Port, Jc, Jmin, Jmax, S1, S2 int
	H1, H2, H3, H4 int64
	Peers []AWGPeerEntry
}

func BuildAWGServerConfig(awg *config.StealthAWGConfig, users []models.User, peers []models.WireGuardPeer) ([]byte, error) {
	if awg == nil || !awg.Enabled { return nil, fmt.Errorf("awg not enabled") }
	m := map[uint]models.WireGuardPeer{}
	for _, p := range peers { m[p.UserID] = p }
	entries := []AWGPeerEntry{}
	for _, u := range users {
		p, ok := m[u.ID]; if !ok { continue }
		entries = append(entries, AWGPeerEntry{u.ID, u.Email, p.PublicKey, p.PresharedKey, strings.TrimSuffix(p.ClientIP, "/32")})
	}
	data := awgTemplateData{awg.PrivateKey, awg.ServerAddressOrDefault(), awg.PortOrDefault(), awg.Jc, awg.Jmin, awg.Jmax, awg.S1, awg.S2, awg.H1, awg.H2, awg.H3, awg.H4, entries}
	content, err := templateFS.ReadFile("templates/awg0.conf.tmpl"); if err != nil { return nil, err }
	tmpl, err := template.New("awg0.conf.tmpl").Parse(string(content)); if err != nil { return nil, err }
	var buf bytes.Buffer
	return buf.Bytes(), tmpl.Execute(&buf, data)
}

func BuildAWGClientConfig(host string, awg *config.StealthAWGConfig, peer *models.WireGuardPeer) string {
	if awg == nil || !awg.Enabled || peer == nil { return "" }
	var b strings.Builder
	b.WriteString("[Interface]\nPrivateKey = " + peer.PrivateKey + "\nAddress = " + peer.ClientIP + "\nDNS = 1.1.1.1, 8.8.8.8\n")
	if awg.Jc > 0 {
		fmt.Fprintf(&b, "Jc = %d\nJmin = %d\nJmax = %d\nS1 = %d\nS2 = %d\nH1 = %d\nH2 = %d\nH3 = %d\nH4 = %d\n", awg.Jc, awg.Jmin, awg.Jmax, awg.S1, awg.S2, awg.H1, awg.H2, awg.H3, awg.H4)
	}
	b.WriteString("\n[Peer]\nPublicKey = " + awg.PublicKey + "\n")
	if peer.PresharedKey != "" { b.WriteString("PresharedKey = " + peer.PresharedKey + "\n") }
	b.WriteString("AllowedIPs = 0.0.0.0/0, ::/0\n")
	fmt.Fprintf(&b, "Endpoint = %s:%d\nPersistentKeepalive = 25\n", host, awg.PortOrDefault())
	return b.String()
}

func BuildAWGURILink(ini string) string {
	if ini == "" { return "" }
	return "awg://" + base64.URLEncoding.EncodeToString([]byte(ini))
}
