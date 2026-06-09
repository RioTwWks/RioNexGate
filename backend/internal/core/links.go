package core

import (
	"fmt"
	"net/url"

	"proxy-mgr/internal/models"
)

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
	// Simplified vmess URI for MVP
	return fmt.Sprintf("vmess://%s@%s:%d", user.UUID, host, port)
}

func GetClientLink(host string, port int, user models.User, protocol string) string {
	switch protocol {
	case "vmess":
		return buildVMessLink(host, port, user)
	default:
		return buildVLESSLink(host, port, user)
	}
}
