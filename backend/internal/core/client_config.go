package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"rionexgate/internal/config"
	"rionexgate/internal/models"
)

type ClientServer struct {
	Protocol   string            `json:"protocol"`
	Link       string            `json:"link"`
	ID         string            `json:"id"`
	Host       string            `json:"host"`
	Port       int               `json:"port"`
	Encryption string            `json:"encryption,omitempty"`
	Params     map[string]string `json:"transport,omitempty"`
}

type ClientSOCKS5Inbound struct {
	Port int    `json:"port"`
	Auth string `json:"auth"`
}

type ClientInbounds struct {
	SOCKS5 ClientSOCKS5Inbound `json:"socks5"`
}

type ClientDNS struct {
	Servers []string `json:"servers"`
}

type ClientConfigBody struct {
	Servers  []ClientServer `json:"servers"`
	Profiles []LinkProfile  `json:"profiles"`
	Inbounds ClientInbounds `json:"inbounds"`
	DNS      ClientDNS      `json:"dns"`
}

type ClientConfig struct {
	ConfigHash string         `json:"config_hash"`
	Servers    []ClientServer `json:"servers"`
	Profiles   []LinkProfile  `json:"profiles"`
	Inbounds   ClientInbounds `json:"inbounds"`
	DNS        ClientDNS      `json:"dns"`
}

func BuildClientConfig(host string, port int, user models.User, socksPort int, stealth *config.StealthConfig, entry *models.Node, peer *models.WireGuardPeer) (*ClientConfig, error) {
	ep := ResolveClientEndpoint(host, port, user, entry)
	servers := make([]ClientServer, 0, len(SupportedProtocols))
	for _, proto := range SupportedProtocols {
		link := GetClientLink(ep.Host, ep.Port, user, proto, stealth)
		srvPort := ep.Port
		if proto == "vless" {
			srvPort = vlessLinkPort(link, ep.Port)
		}
		srv := ClientServer{
			Protocol: proto,
			Link:     link,
			ID:       user.UUID,
			Host:     ep.Host,
			Port:     srvPort,
		}
		switch proto {
		case "vless":
			srv.Encryption = "none"
			srv.Params = vlessTransportParams(link)
		case "vmess":
			srv.Params = map[string]string{"net": "tcp", "type": "none"}
		case "trojan":
			srv.Params = map[string]string{"type": "tcp", "security": "tls"}
		}
		servers = append(servers, srv)
	}

	if socksPort == 0 {
		socksPort = 10808
	}

	profiles := GetClientLinkProfiles(ep.Host, ep.Port, user, stealth, peer)

	body := ClientConfigBody{
		Servers:  servers,
		Profiles: profiles,
		Inbounds: ClientInbounds{
			SOCKS5: ClientSOCKS5Inbound{Port: socksPort, Auth: "none"},
		},
		DNS: ClientDNS{Servers: []string{"1.1.1.1", "8.8.8.8"}},
	}

	hash, err := ConfigHash(body)
	if err != nil {
		return nil, err
	}

	return &ClientConfig{
		ConfigHash: hash,
		Servers:    body.Servers,
		Profiles:   body.Profiles,
		Inbounds:   body.Inbounds,
		DNS:        body.DNS,
	}, nil
}

// vlessLinkPort extracts the destination port from a VLESS share link.
func vlessLinkPort(link string, fallback int) int {
	u, err := url.Parse(link)
	if err != nil || u.Host == "" {
		return fallback
	}
	if p := u.Port(); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}

// vlessTransportParams derives transport metadata from a VLESS share link for RioNexTunnel.
func vlessTransportParams(link string) map[string]string {
	params := map[string]string{"type": "tcp", "security": "none"}
	if !strings.HasPrefix(link, "vless://") {
		return params
	}
	rest := strings.TrimPrefix(link, "vless://")
	if idx := strings.Index(rest, "?"); idx >= 0 {
		q, err := url.ParseQuery(rest[idx+1:])
		if err == nil {
			if t := q.Get("type"); t != "" {
				params["type"] = t
			}
			if s := q.Get("security"); s != "" {
				params["security"] = s
			}
			if f := q.Get("flow"); f != "" {
				params["flow"] = f
			}
		}
	}
	return params
}

func ConfigHash(body ClientConfigBody) (string, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func (c *ClientConfig) JSON() ([]byte, error) {
	return json.Marshal(c)
}
