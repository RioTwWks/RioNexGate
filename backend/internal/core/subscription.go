package core

import (
	"encoding/base64"
	"strings"

	"rionexgate/internal/config"
	"rionexgate/internal/models"
)

func BuildSubscriptionLinks(host string, port int, user models.User, stealth *config.StealthConfig, entry *models.Node, exits []models.UserExitAssignment, multihop *config.MultihopConfig, peer *models.WireGuardPeer, coreType string) []string {
	ep := ResolveClientEndpoint(host, port, user, entry)
	var primaryExit *models.Node
	if len(exits) == 1 {
		primaryExit = &exits[0].Node
	} else if len(exits) > 1 {
		primaryExit = &exits[0].Node
	}
	_ = primaryExit
	profiles := GetClientLinkProfilesForExits(ep.Host, ep.Port, user, stealth, peer, multihop, exits)
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
	// Only append legacy protocols when there is a single (or no) exit — multi-exit subscriptions
	// stay VLESS-profile only to avoid N×protocol noise.
	if !IsSkadiCore(coreType) && LegacyProtocolsInSubscription(stealth) && len(exits) <= 1 {
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

func BuildSubscriptionBase64(host string, port int, user models.User, stealth *config.StealthConfig, entry *models.Node, exits []models.UserExitAssignment, multihop *config.MultihopConfig, peer *models.WireGuardPeer, coreType string) string {
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(BuildSubscriptionLinks(host, port, user, stealth, entry, exits, multihop, peer, coreType), "\n")))
}

func BuildSubscriptionBase64Graceful(host string, port int, user models.User, stealth *config.StealthConfig, entry *models.Node, exits []models.UserExitAssignment, multihop *config.MultihopConfig, peer *models.WireGuardPeer, coreType string) string {
	links := BuildSubscriptionLinks(host, port, user, stealth, entry, exits, multihop, peer, coreType)
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
