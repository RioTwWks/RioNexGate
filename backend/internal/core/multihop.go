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
	Tag   string
	Node  models.Node
	Creds models.NodeCredentials
	Chain bool
}

// MultihopRouting maps inbound client emails to an exit outbound tag.
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

// MultihopChainActive reports whether the user should use entry→exit chaining.
// When true, XHTTP profiles are omitted from client-facing output: Vision/TCP matches
// the RU→EU relay transport and works reliably in RioNexTunnel on mobile.
func MultihopChainActive(multihop *config.MultihopConfig, exit *models.Node) bool {
	return multihop != nil && multihop.IsEntryNode() && exit != nil &&
		exit.Active && exit.Role == models.NodeRoleExit
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

// ExpandInboundUsers turns panel users + exit assignments into inbound VLESS clients.
// One panel user with N exits becomes N inbound clients (different UUID/email per exit)
// so the entry core can route each country selection independently.
func ExpandInboundUsers(users []models.User, assignments []models.UserExitAssignment) []models.User {
	byUser := map[uint][]models.UserExitAssignment{}
	for _, a := range assignments {
		byUser[a.UserExit.UserID] = append(byUser[a.UserExit.UserID], a)
	}
	out := make([]models.User, 0, len(users)+len(assignments))
	for _, u := range users {
		as := byUser[u.ID]
		if len(as) == 0 {
			out = append(out, u)
			continue
		}
		for _, a := range as {
			client := u
			client.UUID = a.UserExit.UUID
			client.Email = a.UserExit.Email
			out = append(out, client)
		}
	}
	return out
}

// BuildMultihopData builds outbound/routing data for entry-node xray configs.
// Deprecated path: single exit per user via resolveExit. Prefer BuildMultihopDataFromAssignments.
func BuildMultihopData(multihop *config.MultihopConfig, users []models.User, exitNodes []models.Node, resolveExit func(models.User) *models.Node) MultihopData {
	if multihop == nil || !multihop.IsEntryNode() || len(exitNodes) == 0 {
		return MultihopData{}
	}

	var assignments []models.UserExitAssignment
	for _, user := range users {
		exit := resolveExit(user)
		if exit == nil {
			continue
		}
		assignments = append(assignments, models.UserExitAssignment{
			UserExit: models.UserExit{
				UserID: user.ID,
				NodeID: exit.ID,
				UUID:   user.UUID,
				Email:  user.Email,
			},
			Node: *exit,
		})
	}
	return BuildMultihopDataFromAssignments(multihop, assignments)
}

// BuildMultihopDataFromAssignments builds outbounds for every referenced exit and
// routes each assignment email to that exit's chain tag.
func BuildMultihopDataFromAssignments(multihop *config.MultihopConfig, assignments []models.UserExitAssignment) MultihopData {
	if multihop == nil || !multihop.IsEntryNode() || len(assignments) == 0 {
		return MultihopData{}
	}

	outboundByID := make(map[uint]MultihopOutbound)
	routeByTag := make(map[string][]string)

	for _, a := range assignments {
		exit := a.Node
		if !exit.Active || exit.Role != models.NodeRoleExit {
			continue
		}
		if _, ok := outboundByID[exit.ID]; !ok {
			creds := exit.ParsedCredentials().NormalizeForOutbound(exit.Address)
			outboundByID[exit.ID] = MultihopOutbound{
				Tag:   exit.OutboundTag(),
				Node:  exit,
				Creds: creds,
				Chain: true,
			}
		}
		email := a.UserExit.Email
		if email == "" {
			continue
		}
		tag := exit.OutboundTag()
		routeByTag[tag] = append(routeByTag[tag], email)
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
