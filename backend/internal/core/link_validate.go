package core

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// ValidateClientLink checks that a share link conforms to the VLESS/VMess/Trojan URI specs.
func ValidateClientLink(link string) error {
	if link == "" {
		return fmt.Errorf("empty link")
	}
	switch {
	case strings.HasPrefix(link, "vless://"):
		return validateVLESSLink(link)
	case strings.HasPrefix(link, "vmess://"):
		return validateVMessLink(link)
	case strings.HasPrefix(link, "trojan://"):
		return validateTrojanLink(link)
	default:
		return fmt.Errorf("unsupported link scheme")
	}
}

func validateVLESSLink(link string) error {
	u, err := url.Parse(link)
	if err != nil {
		return fmt.Errorf("invalid vless url: %w", err)
	}
	if u.User == nil || u.User.Username() == "" {
		return fmt.Errorf("vless: missing id (uuid)")
	}
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil || host == "" {
		return fmt.Errorf("vless: missing host or port")
	}
	if port == "" {
		return fmt.Errorf("vless: missing port")
	}
	if _, err := strconv.Atoi(port); err != nil {
		return fmt.Errorf("vless: invalid port")
	}

	q := u.Query()
	if enc := q.Get("encryption"); enc == "" {
		return fmt.Errorf("vless: missing encryption")
	}

	transport := q.Get("type")
	if transport == "" {
		transport = "tcp"
	}
	security := q.Get("security")
	if security == "" {
		security = "none"
	}

	switch transport {
	case "ws", "httpupgrade", "splithttp", "xhttp":
		if q.Get("path") == "" {
			return fmt.Errorf("vless: %s transport requires path", transport)
		}
		if transport == "ws" && q.Get("host") == "" && q.Get("sni") == "" {
			return fmt.Errorf("vless: ws transport requires host or sni")
		}
	case "grpc":
		if q.Get("serviceName") == "" {
			return fmt.Errorf("vless: grpc transport requires serviceName")
		}
	}

	switch security {
	case "reality":
		for _, key := range []string{"pbk", "sid", "sni", "fp"} {
			if q.Get(key) == "" {
				return fmt.Errorf("vless: reality security requires %s", key)
			}
		}
	case "tls":
		if q.Get("sni") == "" {
			return fmt.Errorf("vless: tls security requires sni")
		}
	}
	return nil
}

func validateVMessLink(link string) error {
	raw := strings.TrimPrefix(link, "vmess://")
	if raw == "" {
		return fmt.Errorf("vmess: empty payload")
	}
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return fmt.Errorf("vmess: invalid base64 payload: %w", err)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return fmt.Errorf("vmess: invalid json payload: %w", err)
	}
	for _, key := range []string{"v", "id", "add", "port"} {
		val, ok := payload[key]
		if !ok {
			return fmt.Errorf("vmess: missing required field %s", key)
		}
		if s, ok := val.(string); ok && strings.TrimSpace(s) == "" {
			return fmt.Errorf("vmess: empty required field %s", key)
		}
	}
	if v, _ := payload["v"].(string); v != "2" {
		return fmt.Errorf("vmess: unsupported version %q", v)
	}
	netType, _ := payload["net"].(string)
	if netType == "" {
		netType = "tcp"
	}
	switch netType {
	case "ws", "http", "h2":
		if path, _ := payload["path"].(string); strings.TrimSpace(path) == "" {
			return fmt.Errorf("vmess: %s transport requires path", netType)
		}
	}
	if tls, _ := payload["tls"].(string); tls == "tls" {
		if sni, _ := payload["sni"].(string); strings.TrimSpace(sni) == "" {
			if host, _ := payload["host"].(string); strings.TrimSpace(host) == "" {
				return fmt.Errorf("vmess: tls requires sni or host")
			}
		}
	}
	return nil
}

func validateTrojanLink(link string) error {
	u, err := url.Parse(link)
	if err != nil {
		return fmt.Errorf("invalid trojan url: %w", err)
	}
	if u.User == nil || u.User.Username() == "" {
		return fmt.Errorf("trojan: missing password")
	}
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil || host == "" {
		return fmt.Errorf("trojan: missing host or port")
	}
	if port == "" {
		return fmt.Errorf("trojan: missing port")
	}
	if _, err := strconv.Atoi(port); err != nil {
		return fmt.Errorf("trojan: invalid port")
	}
	return nil
}
