package core

import (
	"encoding/base64"
	"strings"

	"rionexgate/internal/config"
	"rionexgate/internal/models"
)

func BuildSubscriptionLinks(host string, port int, user models.User, stealth *config.StealthConfig, entry *models.Node) []string {
	ep := ResolveClientEndpoint(host, port, user, entry)
	links := make([]string, 0)
	for _, p := range GetClientLinkProfiles(ep.Host, ep.Port, user, stealth) {
		if p.Link != "" {
			links = append(links, p.Link)
		}
	}
	for _, proto := range SupportedProtocols {
		if proto == "vless" {
			continue
		}
		link := GetClientLink(ep.Host, ep.Port, user, proto, stealth)
		if link != "" {
			links = append(links, link)
		}
	}
	return links
}

func BuildSubscriptionBase64(host string, port int, user models.User, stealth *config.StealthConfig, entry *models.Node) string {
	links := BuildSubscriptionLinks(host, port, user, stealth, entry)
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(links, "\n")))
}

func BuildSubscriptionBase64Graceful(host string, port int, user models.User, stealth *config.StealthConfig, entry *models.Node) string {
	links := BuildSubscriptionLinks(host, port, user, stealth, entry)
	if len(links) == 0 {
		links = []string{GetClientLink(host, port, user, "vless", stealth)}
	}
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(links, "\n")))
}
