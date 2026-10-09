package core

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"rionexgate/internal/config"
	"rionexgate/internal/models"
)

// SupportedProtocols lists client link protocols available in the UI and API.
var SupportedProtocols = []string{"vless", "vmess", "trojan"}

// LinkProfile describes one stealth transport profile for a user.
type LinkProfile struct {
	Profile   string `json:"profile"`
	Transport string `json:"transport"`
	Priority  int    `json:"priority"`
	Port      int    `json:"port"`
	Tags      string `json:"tags,omitempty"`
	Link      string `json:"link"`
	Config    string `json:"config,omitempty"`
}

func buildVLESSLink(host string, port int, user models.User) string {
	params := url.Values{}
	params.Set("encryption", "none")
	params.Set("type", "tcp")
	params.Set("security", "none")
	fragment := url.PathEscape(user.Email)
	return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
		user.UUID, host, port, params.Encode(), fragment)
}

func profileRemark(base, suffix string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		base = "profile"
	}
	if suffix == "" {
		return url.PathEscape(base)
	}
	return url.PathEscape(base + "-" + suffix)
}

func buildVLESSRealityXHTTPLink(host string, port int, user models.User, stealth *config.StealthConfig, remarkBase string) string {
	params := url.Values{}
	params.Set("encryption", "none")
	params.Set("type", "xhttp")
	params.Set("security", "reality")
	params.Set("sni", stealth.PrimarySNI())
	params.Set("fp", stealth.FingerprintOrDefault())
	params.Set("pbk", stealth.Reality.PublicKey)
	params.Set("sid", stealth.PrimaryShortID())
	params.Set("path", stealth.XHTTP.Path)
	params.Set("mode", stealth.XHTTP.Mode)
	if remarkBase == "" {
		remarkBase = user.Email
	}
	return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
		user.UUID, host, port, params.Encode(), profileRemark(remarkBase, "xhttp"))
}

func buildVLESSRealityVisionLink(host string, port int, user models.User, stealth *config.StealthConfig, remarkBase string) string {
	params := url.Values{}
	params.Set("encryption", "none")
	params.Set("type", "tcp")
	params.Set("flow", "xtls-rprx-vision")
	params.Set("security", "reality")
	params.Set("sni", stealth.PrimarySNI())
	params.Set("fp", stealth.FingerprintOrDefault())
	params.Set("pbk", stealth.Reality.PublicKey)
	params.Set("sid", stealth.PrimaryShortID())
	if remarkBase == "" {
		remarkBase = user.Email
	}
	return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
		user.UUID, host, port, params.Encode(), profileRemark(remarkBase, "vision"))
}

func buildVLESSTLSLink(host string, port int, user models.User, stealth *config.StealthConfig, remarkBase string) string {
	params := url.Values{}
	params.Set("encryption", "none")
	params.Set("type", "tcp")
	params.Set("security", "tls")
	sni := stealth.TLS.SNI
	if sni == "" {
		sni = host
	}
	params.Set("sni", sni)
	params.Set("fp", stealth.FingerprintOrDefault())
	if len(stealth.TLS.ALPN) > 0 {
		params.Set("alpn", strings.Join(stealth.TLS.ALPN, ","))
	}
	if remarkBase == "" {
		remarkBase = user.Email
	}
	return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
		user.UUID, host, port, params.Encode(), profileRemark(remarkBase, "tls"))
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

// primaryVLESSStealthLink returns the lowest-latency VLESS profile when stealth is active.
// Vision/TCP is preferred over XHTTP for multihop and mobile clients; XHTTP remains available
// as a separate subscription profile for DPI-heavy networks.
func primaryVLESSStealthLink(host string, port int, user models.User, stealth *config.StealthConfig) string {
	if stealth == nil || !stealth.IsActive() {
		return buildVLESSLink(host, port, user)
	}
	if stealth.Vision.Enabled {
		return buildVLESSRealityVisionLink(host, stealth.Vision.Port, user, stealth, user.Email)
	}
	if stealth.XHTTP.Enabled {
		return buildVLESSRealityXHTTPLink(host, stealth.XHTTP.Port, user, stealth, user.Email)
	}
	if stealth.TLS.Enabled {
		return buildVLESSTLSLink(host, stealth.TLS.Port, user, stealth, user.Email)
	}
	return buildVLESSLink(host, port, user)
}

// LegacyProtocolsInSubscription reports whether plain VMess/Trojan links belong in a subscription.
// With stealth inbounds (Reality/XHTTP/Vision), legacy TCP links use the wrong port and security.
func LegacyProtocolsInSubscription(stealth *config.StealthConfig) bool {
	return stealth == nil || !stealth.IsActive()
}

// ClientConfigProtocols returns protocols for RioNexTunnel servers[].
func ClientConfigProtocols(stealth *config.StealthConfig) []string {
	if LegacyProtocolsInSubscription(stealth) {
		return SupportedProtocols
	}
	return []string{"vless"}
}

// GetClientLink returns a single client link for the given protocol.
// For vless with stealth enabled, the primary Vision/TCP profile is returned when available.
func GetClientLink(host string, port int, user models.User, protocol string, stealth *config.StealthConfig) string {
	switch protocol {
	case "vmess":
		return buildVMessLink(host, port, user)
	case "trojan":
		return buildTrojanLink(host, port, user)
	default:
		return primaryVLESSStealthLink(host, port, user, stealth)
	}
}

// GetClientLinkProfiles returns all available stealth profiles for a user.
// remarkBase overrides the URI fragment prefix (e.g. country label "NL"); empty uses email.
func GetClientLinkProfiles(host string, port int, user models.User, stealth *config.StealthConfig, peer *models.WireGuardPeer, multihop *config.MultihopConfig, exit *models.Node, remarkBase string) []LinkProfile {
	if remarkBase == "" {
		remarkBase = user.Email
	}
	if stealth == nil || (!stealth.IsActive() && !stealth.AWGActive()) {
		return []LinkProfile{{Profile: "legacy-tcp", Transport: "tcp", Priority: 1, Port: port, Link: buildVLESSLink(host, port, user)}}
	}
	var profiles []LinkProfile
	priority := 1
	if stealth.IsActive() {
		// Vision/TCP first: lowest latency for multihop and mobile; matches RU→EU relay transport.
		if stealth.Vision.Enabled {
			profiles = append(profiles, LinkProfile{
				Profile: "vision-tcp-primary", Transport: "tcp", Priority: priority,
				Port: stealth.Vision.Port, Tags: "vision-tcp-primary,low-latency",
				Link: buildVLESSRealityVisionLink(host, stealth.Vision.Port, user, stealth, remarkBase),
			})
			priority++
		}
		if stealth.XHTTP.Enabled {
			profiles = append(profiles, LinkProfile{
				Profile: "xhttp-anti-dpi", Transport: "xhttp", Priority: priority,
				Port: stealth.XHTTP.Port, Tags: "xhttp-anti-dpi",
				Link: buildVLESSRealityXHTTPLink(host, stealth.XHTTP.Port, user, stealth, remarkBase),
			})
			priority++
		}
		if stealth.TLS.Enabled {
			profiles = append(profiles, LinkProfile{
				Profile: "tls-mobile", Transport: "tcp", Priority: priority,
				Port: stealth.TLS.Port, Tags: "tls-mobile,mux-hint",
				Link: buildVLESSTLSLink(host, stealth.TLS.Port, user, stealth, remarkBase),
			})
			priority++
		}
	}
	if stealth.AWGActive() && peer != nil {
		ini := BuildAWGClientConfig(host, &stealth.AWG, peer)
		profiles = append(profiles, LinkProfile{Profile: "awg-udp-reserve", Transport: "awg", Priority: priority, Port: stealth.AWG.PortOrDefault(), Tags: "awg-reserve,udp", Link: BuildAWGURILink(ini), Config: ini})
	}
	if len(profiles) == 0 {
		return []LinkProfile{{Profile: "legacy-tcp", Transport: "tcp", Priority: 1, Port: port, Link: buildVLESSLink(host, port, user)}}
	}
	return omitXHTTPProfilesForMultihop(profiles, multihop, exit)
}

// GetClientLinkProfilesForExits emits transport profiles for each exit assignment so a
// single subscription can list countries (NL, DE, …) as selectable servers.
func GetClientLinkProfilesForExits(host string, port int, user models.User, stealth *config.StealthConfig, peer *models.WireGuardPeer, multihop *config.MultihopConfig, assignments []models.UserExitAssignment) []LinkProfile {
	if len(assignments) == 0 {
		return GetClientLinkProfiles(host, port, user, stealth, peer, multihop, nil, user.Email)
	}
	var all []LinkProfile
	priority := 1
	multi := len(assignments) > 1
	for _, a := range assignments {
		client := user
		client.UUID = a.UserExit.UUID
		client.Email = a.UserExit.Email
		label := models.ExitDisplayLabel(a.Node)
		remark := user.Email
		if multi {
			remark = label
		}
		exit := a.Node
		profiles := GetClientLinkProfiles(host, port, client, stealth, peer, multihop, &exit, remark)
		for _, p := range profiles {
			p.Priority = priority
			if multi {
				if p.Tags != "" {
					p.Tags = p.Tags + "," + models.ExitSlug(a.Node)
				} else {
					p.Tags = models.ExitSlug(a.Node)
				}
				p.Profile = models.ExitSlug(a.Node) + "-" + p.Profile
			}
			all = append(all, p)
			priority++
		}
	}
	return all
}

// omitXHTTPProfilesForMultihop removes XHTTP from client output when an entry→exit chain is active.
// RioNexTunnel subscription mode remembers the last selected server; without this, users can stay
// stuck on XHTTP which often fails to connect on mobile even though Vision/TCP works.
func omitXHTTPProfilesForMultihop(profiles []LinkProfile, multihop *config.MultihopConfig, exit *models.Node) []LinkProfile {
	if !MultihopChainActive(multihop, exit) {
		return profiles
	}
	filtered := make([]LinkProfile, 0, len(profiles))
	priority := 1
	for _, p := range profiles {
		if p.Transport == "xhttp" {
			continue
		}
		p.Priority = priority
		priority++
		filtered = append(filtered, p)
	}
	return filtered
}

// FormatSubscriptionLinks joins profile links with newlines for base64 subscription encoding.
func FormatSubscriptionLinks(profiles []LinkProfile) string {
	lines := make([]string, 0, len(profiles))
	for _, p := range profiles {
		if p.Link != "" {
			lines = append(lines, p.Link)
		}
	}
	return strings.Join(lines, "\n")
}

// EncodeSubscription returns base64-encoded subscription content.
func EncodeSubscription(profiles []LinkProfile) string {
	return base64.StdEncoding.EncodeToString([]byte(FormatSubscriptionLinks(profiles)))
}
