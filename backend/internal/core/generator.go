package core

import (
	"bytes"
	"embed"
	"text/template"

	"proxy-mgr/internal/models"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

type templateData struct {
	ListenPort int
	APIAddress string
	Users      []models.User
}

func renderTemplate(name string, data templateData) ([]byte, error) {
	content, err := templateFS.ReadFile("templates/" + name)
	if err != nil {
		return nil, err
	}
	tmpl, err := template.New(name).Parse(string(content))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func generateXrayConfig(listenPort int, apiAddress string, users []models.User) ([]byte, error) {
	return renderTemplate("xray.json.tmpl", templateData{
		ListenPort: listenPort,
		APIAddress: apiAddress,
		Users:      users,
	})
}

func generateSingboxConfig(listenPort int, apiAddress string, users []models.User) ([]byte, error) {
	return renderTemplate("singbox.json.tmpl", templateData{
		ListenPort: listenPort,
		APIAddress: apiAddress,
		Users:      users,
	})
}
