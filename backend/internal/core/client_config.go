package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

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
	Inbounds ClientInbounds `json:"inbounds"`
	DNS      ClientDNS      `json:"dns"`
}

type ClientConfig struct {
	ConfigHash string         `json:"config_hash"`
	Servers    []ClientServer `json:"servers"`
	Inbounds   ClientInbounds `json:"inbounds"`
	DNS        ClientDNS      `json:"dns"`
}

func BuildClientConfig(host string, port int, user models.User, socksPort int) (*ClientConfig, error) {
	servers := make([]ClientServer, 0, len(SupportedProtocols))
	for _, proto := range SupportedProtocols {
		link := GetClientLink(host, port, user, proto)
		srv := ClientServer{
			Protocol: proto,
			Link:     link,
			ID:       user.UUID,
			Host:     host,
			Port:     port,
		}
		switch proto {
		case "vless":
			srv.Encryption = "none"
			srv.Params = map[string]string{"type": "tcp", "security": "none"}
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

	body := ClientConfigBody{
		Servers: servers,
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
		Inbounds:   body.Inbounds,
		DNS:        body.DNS,
	}, nil
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
