package core

import (
	"bytes"
	"embed"
	"strings"
	"text/template"

	"rionexgate/internal/config"
	"rionexgate/internal/models"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

type templateData struct {
	ListenPort  int
	APIAddress  string
	Users       []models.User
	Stealth     *config.StealthConfig
	Multihop    MultihopData
	InboundTags []string
}

func renderTemplate(name string, data templateData) ([]byte, error) {
	content, err := templateFS.ReadFile("templates/" + name)
	if err != nil {
		return nil, err
	}
	funcMap := template.FuncMap{
		"jsonStrings":   jsonStringList,
		"stealthSNI":    stealthPrimarySNI,
		"stealthActive": stealthIsActive,
	}
	tmpl, err := template.New(name).Funcs(funcMap).Parse(string(content))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func stealthIsActive(s *config.StealthConfig) bool {
	return s != nil && s.IsActive()
}

func stealthPrimarySNI(s *config.StealthConfig) string {
	if s == nil {
		return ""
	}
	return s.PrimarySNI()
}

func jsonStringList(items []string) string {
	quoted := make([]string, len(items))
	for i, item := range items {
		quoted[i] = `"` + item + `"`
	}
	return strings.Join(quoted, ", ")
}

func generateXrayConfig(listenPort int, apiAddress string, users []models.User, stealth *config.StealthConfig, multihop MultihopData) ([]byte, error) {
	return renderTemplate("xray.json.tmpl", templateData{
		ListenPort:  listenPort,
		APIAddress:  apiAddress,
		Users:       users,
		Stealth:     stealth,
		Multihop:    multihop,
		InboundTags: CollectInboundTags(stealth, "vless-in"),
	})
}

func generateSingboxConfig(listenPort int, apiAddress string, users []models.User, stealth *config.StealthConfig, multihop MultihopData) ([]byte, error) {
	return renderTemplate("singbox.json.tmpl", templateData{
		ListenPort: listenPort,
		APIAddress: apiAddress,
		Users:      users,
		Stealth:    stealth,
		Multihop:   multihop,
	})
}
