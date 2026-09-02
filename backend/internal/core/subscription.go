package core

import (
	"encoding/base64"
	"strings"

	"rionexgate/internal/models"
)

func BuildSubscriptionLinks(host string, port int, user models.User) []string {
	links := make([]string, 0, len(SupportedProtocols))
	for _, proto := range SupportedProtocols {
		link := GetClientLink(host, port, user, proto)
		if link != "" {
			links = append(links, link)
		}
	}
	return links
}

func BuildSubscriptionBase64(host string, port int, user models.User) string {
	links := BuildSubscriptionLinks(host, port, user)
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(links, "\n")))
}

func BuildSubscriptionBase64Graceful(host string, port int, user models.User) string {
	links := BuildSubscriptionLinks(host, port, user)
	if len(links) == 0 {
		links = []string{GetClientLink(host, port, user, "vless")}
	}
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(links, "\n")))
}
