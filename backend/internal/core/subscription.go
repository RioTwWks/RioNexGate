package core
import ("encoding/base64"; "strings"; "rionexgate/internal/config"; "rionexgate/internal/models")
func BuildSubscriptionLinks(host string, port int, user models.User, stealth *config.StealthConfig, entry *models.Node, peer *models.WireGuardPeer) []string {
	ep := ResolveClientEndpoint(host, port, user, entry); links := []string{}
	for _, p := range GetClientLinkProfiles(ep.Host, ep.Port, user, stealth, peer) { if p.Link != "" { links = append(links, p.Link) } }
	for _, proto := range SupportedProtocols { if proto == "vless" { continue }; if l := GetClientLink(ep.Host, ep.Port, user, proto, stealth); l != "" { links = append(links, l) } }
	return links
}
func BuildSubscriptionBase64(host string, port int, user models.User, stealth *config.StealthConfig, entry *models.Node, peer *models.WireGuardPeer) string {
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(BuildSubscriptionLinks(host, port, user, stealth, entry, peer), "\n")))
}
func BuildSubscriptionBase64Graceful(host string, port int, user models.User, stealth *config.StealthConfig, entry *models.Node, peer *models.WireGuardPeer) string {
	links := BuildSubscriptionLinks(host, port, user, stealth, entry, peer)
	if len(links) == 0 { links = []string{GetClientLink(host, port, user, "vless", stealth)} }
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(links, "\n")))
}
