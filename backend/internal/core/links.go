package core

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"rionexgate/internal/models"
)

// SupportedProtocols lists client link protocols available in the UI and API.
var SupportedProtocols = []string{"vless", "vmess", "trojan"}

func buildVLESSLink(host string, port int, user models.User) string {
	params := url.Values{}
	params.Set("encryption", "none")
	params.Set("type", "tcp")
	params.Set("security", "none")
	fragment := url.PathEscape(user.Email)
	return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
		user.UUID, host, port, params.Encode(), fragment)
}

func buildVMessLink(host string, port int, user models.User) string {
	payload := map[string]string{
		"v":    "2",
		"ps":   user.Email,
		"add":  host,
		"port": strconv.Itoa(port),
		"id":   user.UUID,
		"aid":  "0",
		"net":  "tcp",
		"type": "none",
		"host": "",
		"path": "",
		"tls":  "",
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Sprintf("vmess://%s@%s:%d", user.UUID, host, port)
	}
	return "vmess://" + base64.StdEncoding.EncodeToString(raw)
}

func buildTrojanLink(host string, port int, user models.User) string {
	params := url.Values{}
	params.Set("allowInsecure", "1")
	params.Set("type", "tcp")
	if host != "" {
		params.Set("sni", host)
	}
	fragment := url.PathEscape(user.Email)
	return fmt.Sprintf("trojan://%s@%s:%d?%s#%s",
		user.UUID, host, port, params.Encode(), fragment)
}

func GetClientLink(host string, port int, user models.User, protocol string) string {
	switch protocol {
	case "vmess":
		return buildVMessLink(host, port, user)
	case "trojan":
		return buildTrojanLink(host, port, user)
	default:
		return buildVLESSLink(host, port, user)
	}
}
