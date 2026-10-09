package core

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
	"strings"
	"text/template"

	"rionexgate/internal/config"
	"rionexgate/internal/models"
)

// skadiTemplateData is passed to skadi.toml.tmpl.
type skadiTemplateData struct {
	ListenAddrs      []string
	Users            []models.User
	Stealth          *config.StealthConfig
	UseReality       bool
	UseXHTTP         bool
	APIAddress       string
	APIListenAddress string
	APIToken         string
	MetricsAddress   string
}

// skadiListenPort picks the single inbound port SkadiCore binds to.
// Skadi listens on one port (unlike Xray dual XHTTP+Vision inbounds).
func skadiListenPort(listenPort int, stealth *config.StealthConfig) int {
	if stealth != nil && stealth.IsActive() {
		if stealth.XHTTP.Enabled && stealth.XHTTP.Port > 0 {
			return stealth.XHTTP.Port
		}
		if stealth.Vision.Enabled && stealth.Vision.Port > 0 {
			return stealth.Vision.Port
		}
		if stealth.TLS.Enabled && stealth.TLS.Port > 0 {
			return stealth.TLS.Port
		}
	}
	if listenPort > 0 {
		return listenPort
	}
	return 443
}

func skadiListenAddrs(port int) []string {
	return []string{
		fmt.Sprintf("0.0.0.0:%d", port),
		fmt.Sprintf("[::]:%d", port),
	}
}

// skadiAPIListenAddress converts backend-facing api_address into a loopback
// bind for Skadi gRPC (Skadi requires 127.0.0.1 or ::1).
func skadiAPIListenAddress(apiAddress string) string {
	if apiAddress == "" {
		return ""
	}
	host, port, err := net.SplitHostPort(apiAddress)
	if err != nil {
		return apiAddress
	}
	switch host {
	case "host.docker.internal", "0.0.0.0", "":
		return "127.0.0.1:" + port
	case "::", "[::]":
		return "[::1]:" + port
	default:
		return apiAddress
	}
}

func generateSkadiConfig(listenPort int, skadiCfg config.SkadiConfig, users []models.User, stealth *config.StealthConfig) ([]byte, error) {
	port := skadiListenPort(listenPort, stealth)
	useReality := stealth != nil && stealth.IsActive() && stealth.Reality.PrivateKey != ""
	useXHTTP := useReality && stealth.XHTTP.Enabled
	data := skadiTemplateData{
		ListenAddrs:      skadiListenAddrs(port),
		Users:            users,
		Stealth:          stealth,
		UseReality:       useReality,
		UseXHTTP:         useXHTTP,
		APIAddress:       skadiCfg.APIAddress,
		APIListenAddress: skadiAPIListenAddress(skadiCfg.APIAddress),
		APIToken:         skadiCfg.APIToken,
		MetricsAddress:   skadiCfg.MetricsAddress,
	}
	if data.APIAddress != "" && data.APIToken == "" {
		data.APIToken = "rionexgate-skadi-api"
	}
	return renderSkadiTemplate(data)
}

func renderSkadiTemplate(data skadiTemplateData) ([]byte, error) {
	content, err := templateFS.ReadFile("templates/skadi.toml.tmpl")
	if err != nil {
		return nil, err
	}
	funcMap := template.FuncMap{
		"jsonString":    jsonString,
		"jsonStrings":   jsonStringList,
		"stealthSNI":    stealthPrimarySNI,
		"stealthActive": stealthIsActive,
	}
	tmpl, err := template.New("skadi.toml.tmpl").Funcs(funcMap).Parse(string(content))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// buildVLESSRealityTCPLink builds VLESS+Reality over TCP without xtls-rprx-vision.
// SkadiCore rejects Vision flow; this profile matches Skadi TCP+REALITY inbound.
func buildVLESSRealityTCPLink(host string, port int, user models.User, stealth *config.StealthConfig) string {
	params := url.Values{}
	params.Set("encryption", "none")
	params.Set("type", "tcp")
	params.Set("security", "reality")
	params.Set("sni", stealth.PrimarySNI())
	params.Set("fp", stealth.FingerprintOrDefault())
	params.Set("pbk", stealth.Reality.PublicKey)
	params.Set("sid", stealth.PrimaryShortID())
	fragment := url.PathEscape(user.Email + "-reality-tcp")
	return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
		user.UUID, host, port, params.Encode(), fragment)
}

// adaptProfilesForSkadi rewrites client profiles for SkadiCore capabilities:
// no Vision flow; single listen port; Reality±XHTTP (or plain TCP).
func adaptProfilesForSkadi(host string, listenPort int, user models.User, stealth *config.StealthConfig, profiles []LinkProfile) []LinkProfile {
	port := skadiListenPort(listenPort, stealth)
	if stealth == nil || !stealth.IsActive() {
		return []LinkProfile{{
			Profile: "skadi-tcp", Transport: "tcp", Priority: 1, Port: port,
			Link: buildVLESSLink(host, port, user),
		}}
	}

	out := make([]LinkProfile, 0, 3)
	priority := 1

	if stealth.XHTTP.Enabled {
		out = append(out, LinkProfile{
			Profile: "skadi-xhttp", Transport: "xhttp", Priority: priority,
			Port: port, Tags: "skadi,xhttp-anti-dpi",
			Link: buildVLESSRealityXHTTPLink(host, port, user, stealth, user.Email),
		})
		priority++
	}

	out = append(out, LinkProfile{
		Profile: "skadi-reality-tcp", Transport: "tcp", Priority: priority,
		Port: port, Tags: "skadi,reality-tcp",
		Link: buildVLESSRealityTCPLink(host, port, user, stealth),
	})
	priority++

	for _, p := range profiles {
		if p.Transport == "awg" {
			p.Priority = priority
			priority++
			out = append(out, p)
		}
	}
	return out
}

// IsSkadiCore reports whether the active core type is SkadiCore.
func IsSkadiCore(coreType string) bool {
	switch strings.ToLower(strings.TrimSpace(coreType)) {
	case "skadi", "skadicore", "skadi-core":
		return true
	default:
		return false
	}
}

// NormalizeCoreType maps aliases to canonical core type names.
func NormalizeCoreType(t string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "xray":
		return "xray", nil
	case "sing-box", "singbox":
		return "sing-box", nil
	case "skadi", "skadicore", "skadi-core":
		return "skadi", nil
	default:
		return "", fmt.Errorf("unsupported core type: %s", t)
	}
}
