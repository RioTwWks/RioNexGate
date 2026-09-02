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

func buildVLESSRealityXHTTPLink(host string, port int, user models.User, stealth *config.StealthConfig) string {
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
	fragment := url.PathEscape(user.Email + "-xhttp")
	return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
		user.UUID, host, port, params.Encode(), fragment)
}

func buildVLESSRealityVisionLink(host string, port int, user models.User, stealth *config.StealthConfig) string {
	params := url.Values{}
	params.Set("encryption", "none")
	params.Set("type", "tcp")
	params.Set("flow", "xtls-rprx-vision")
	params.Set("security", "reality")
	params.Set("sni", stealth.PrimarySNI())
	params.Set("fp", stealth.FingerprintOrDefault())
	params.Set("pbk", stealth.Reality.PublicKey)
	params.Set("sid", stealth.PrimaryShortID())
	fragment := url.PathEscape(user.Email + "-vision")
	return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
		user.UUID, host, port, params.Encode(), fragment)
}

func buildVLESSTLSLink(host string, port int, user models.User, stealth *config.StealthConfig) string {
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
	fragment := url.PathEscape(user.Email + "-tls")
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

// GetClientLink returns a single client link for the given protocol.
// For vless with stealth enabled, the primary XHTTP profile is returned.
func GetClientLink(host string, port int, user models.User, protocol string, stealth *config.StealthConfig) string {
	switch protocol {
	case "vmess":
		return buildVMessLink(host, port, user)
	case "trojan":
		return buildTrojanLink(host, port, user)
	default:
		if stealth != nil && stealth.IsActive() && stealth.XHTTP.Enabled {
			return buildVLESSRealityXHTTPLink(host, stealth.XHTTP.Port, user, stealth)
		}
		return buildVLESSLink(host, port, user)
	}
}

// GetClientLinkProfiles returns all available stealth profiles for a user.
func GetClientLinkProfiles(host string, port int, user models.User, stealth *config.StealthConfig, peer *models.WireGuardPeer) []LinkProfile {
	if stealth == nil || (!stealth.IsActive() && !stealth.AWGActive()) {
		return []LinkProfile{{Profile: "legacy-tcp", Transport: "tcp", Priority: 1, Port: port, Link: buildVLESSLink(host, port, user)}}
	}
	var profiles []LinkProfile; priority := 1
	if stealth.IsActive() {
		if stealth.XHTTP.Enabled { profiles = append(profiles, LinkProfile{Profile: "xhttp-primary", Transport: "xhttp", Priority: priority, Port: stealth.XHTTP.Port, Tags: "xhttp-primary", Link: buildVLESSRealityXHTTPLink(host, stealth.XHTTP.Port, user, stealth)}); priority++ }
		if stealth.Vision.Enabled { profiles = append(profiles, LinkProfile{Profile: "vision-ios-fallback", Transport: "tcp", Priority: priority, Port: stealth.Vision.Port, Tags: "vision-ios-fallback", Link: buildVLESSRealityVisionLink(host, stealth.Vision.Port, user, stealth)}); priority++ }
		if stealth.TLS.Enabled { profiles = append(profiles, LinkProfile{Profile: "tls-mobile", Transport: "tcp", Priority: priority, Port: stealth.TLS.Port, Tags: "tls-mobile,mux-hint", Link: buildVLESSTLSLink(host, stealth.TLS.Port, user, stealth)}); priority++ }
	}
	if stealth.AWGActive() && peer != nil {
		ini := BuildAWGClientConfig(host, &stealth.AWG, peer)
		profiles = append(profiles, LinkProfile{Profile: "awg-udp-reserve", Transport: "awg", Priority: priority, Port: stealth.AWG.PortOrDefault(), Tags: "awg-reserve,udp", Link: BuildAWGURILink(ini), Config: ini})
	}
	if len(profiles) == 0 { return []LinkProfile{{Profile: "legacy-tcp", Transport: "tcp", Priority: 1, Port: port, Link: buildVLESSLink(host, port, user)}} }
	return profiles
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
