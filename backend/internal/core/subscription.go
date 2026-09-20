package core

import (
	"encoding/base64"
	"strings"

	"rionexgate/internal/config"
	"rionexgate/internal/models"
)

func BuildSubscriptionLinks(host string, port int, user models.User, stealth *config.StealthConfig, entry, exit *models.Node, multihop *config.MultihopConfig, peer *models.WireGuardPeer, coreType string) []string {
	ep := ResolveClientEndpoint(host, port, user, entry)
	profiles := GetClientLinkProfiles(ep.Host, ep.Port, user, stealth, peer, multihop, exit)
	if IsSkadiCore(coreType) {
		profiles = adaptProfilesForSkadi(ep.Host, ep.Port, user, stealth, profiles)
	}
	links := []string{}
	for _, p := range profiles {
		if p.Link != "" {
			links = append(links, p.Link)
		}
	}
	// Plain VMess/Trojan links target legacy TCP inbounds only; with stealth they use the wrong
	// port/security and cause client auto-fallback delays (high reported ping, failed connects).
	// SkadiCore is VLESS-only for panel-managed inbounds.
	if !IsSkadiCore(coreType) && LegacyProtocolsInSubscription(stealth) {
		for _, proto := range SupportedProtocols {
			if proto == "vless" {
				continue
			}
			if l := GetClientLink(ep.Host, ep.Port, user, proto, stealth); l != "" {
				links = append(links, l)
			}
		}
	}
	return links
}

func BuildSubscriptionBase64(host string, port int, user models.User, stealth *config.StealthConfig, entry, exit *models.Node, multihop *config.MultihopConfig, peer *models.WireGuardPeer, coreType string) string {
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(BuildSubscriptionLinks(host, port, user, stealth, entry, exit, multihop, peer, coreType), "\n")))
}

func BuildSubscriptionBase64Graceful(host string, port int, user models.User, stealth *config.StealthConfig, entry, exit *models.Node, multihop *config.MultihopConfig, peer *models.WireGuardPeer, coreType string) string {
	links := BuildSubscriptionLinks(host, port, user, stealth, entry, exit, multihop, peer, coreType)
	if len(links) == 0 {
		if IsSkadiCore(coreType) {
			profiles := adaptProfilesForSkadi(host, port, user, stealth, nil)
			if len(profiles) > 0 {
				links = []string{profiles[0].Link}
			}
		} else {
			links = []string{GetClientLink(host, port, user, "vless", stealth)}
		}
	}
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(links, "\n")))
}
