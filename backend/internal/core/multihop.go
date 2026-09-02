package core

import (
	"rionexgate/internal/config"
	"rionexgate/internal/models"
)

// ClientEndpoint is the host/port clients connect to (entry node).
type ClientEndpoint struct {
	Host string
	Port int
}

// MultihopOutbound describes one exit relay outbound in xray config.
type MultihopOutbound struct {
	Tag     string
	Node    models.Node
	Creds   models.NodeCredentials
	Chain   bool
}

// MultihopRouting maps user emails to an exit outbound tag.
type MultihopRouting struct {
	UserEmails  []string
	OutboundTag string
}

// MultihopData is passed to the xray template for chain generation.
type MultihopData struct {
	Enabled   bool
	Outbounds []MultihopOutbound
	Routings  []MultihopRouting
}

// ResolveClientEndpoint returns the entry host/port for client links.
// Exit nodes are never exposed to clients.
func ResolveClientEndpoint(publicHost string, listenPort int, user models.User, entry *models.Node) ClientEndpoint {
	if entry != nil {
		port := entry.Port
		if port <= 0 {
			port = listenPort
		}
		return ClientEndpoint{Host: entry.Address, Port: port}
	}
	return ClientEndpoint{Host: publicHost, Port: listenPort}
}

// BuildMultihopData builds outbound/routing data for entry-node xray configs.
func BuildMultihopData(multihop *config.MultihopConfig, users []models.User, exitNodes []models.Node, resolveExit func(models.User) *models.Node) MultihopData {
	if multihop == nil || !multihop.IsEntryNode() || len(exitNodes) == 0 {
		return MultihopData{}
	}

	outboundByID := make(map[uint]MultihopOutbound)
	routeByTag := make(map[string][]string)

	for _, user := range users {
		exit := resolveExit(user)
		if exit == nil {
			continue
		}
		if _, ok := outboundByID[exit.ID]; !ok {
			outboundByID[exit.ID] = MultihopOutbound{
				Tag:   exit.OutboundTag(),
				Node:  *exit,
				Creds: exit.ParsedCredentials(),
				Chain: true,
			}
		}
		routeByTag[exit.OutboundTag()] = append(routeByTag[exit.OutboundTag()], user.Email)
	}

	if len(outboundByID) == 0 {
		return MultihopData{}
	}

	data := MultihopData{Enabled: true}
	for _, ob := range outboundByID {
		data.Outbounds = append(data.Outbounds, ob)
	}
	for tag, emails := range routeByTag {
		data.Routings = append(data.Routings, MultihopRouting{
			UserEmails:  emails,
			OutboundTag: tag,
		})
	}
	return data
}

// CollectInboundTags returns inbound tags for routing rules.
func CollectInboundTags(stealth *config.StealthConfig, legacyTag string) []string {
	if stealth != nil && stealth.IsActive() {
		var tags []string
		if stealth.XHTTP.Enabled && stealth.XHTTP.Tag != "" {
			tags = append(tags, stealth.XHTTP.Tag)
		}
		if stealth.Vision.Enabled && stealth.Vision.Tag != "" {
			tags = append(tags, stealth.Vision.Tag)
		}
		if stealth.TLS.Enabled && stealth.TLS.Tag != "" {
			tags = append(tags, stealth.TLS.Tag)
		}
		if len(tags) > 0 {
			return tags
		}
	}
	return []string{legacyTag}
}
