package core

import (
	"encoding/base64"
	"testing"

	"rionexgate/internal/models"
)

func TestValidateClientLinkVLESS(t *testing.T) {
	user := models.User{UUID: "550e8400-e29b-41d4-a716-446655440000", Email: "test@example.com"}
	link := GetClientLink("example.com", 443, user, "vless", nil)
	if err := ValidateClientLink(link); err != nil {
		t.Fatalf("valid vless link rejected: %v", err)
	}
}

func TestValidateClientLinkVMess(t *testing.T) {
	user := models.User{UUID: "550e8400-e29b-41d4-a716-446655440000", Email: "test@example.com"}
	link := GetClientLink("example.com", 443, user, "vmess", nil)
	if err := ValidateClientLink(link); err != nil {
		t.Fatalf("valid vmess link rejected: %v", err)
	}
}

func TestValidateClientLinkTrojan(t *testing.T) {
	user := models.User{UUID: "550e8400-e29b-41d4-a716-446655440000", Email: "test@example.com"}
	link := GetClientLink("example.com", 443, user, "trojan", nil)
	if err := ValidateClientLink(link); err != nil {
		t.Fatalf("valid trojan link rejected: %v", err)
	}
}

func TestValidateStealthProfileLinks(t *testing.T) {
	user := models.User{UUID: "uuid-1", Email: "user@test.com"}
	profiles := GetClientLinkProfiles("host.example", 443, user, testStealthConfig())
	for _, p := range profiles {
		if err := ValidateClientLink(p.Link); err != nil {
			t.Fatalf("profile %s link invalid: %v", p.Profile, err)
		}
		if p.Priority < 1 || p.Transport == "" || p.Tags == "" {
			t.Fatalf("invalid profile metadata: %+v", p)
		}
	}
}

func TestValidateClientLinkRejectsInvalid(t *testing.T) {
	cases := []struct {
		name string
		link string
	}{
		{"empty", ""},
		{"bad scheme", "socks5://user@host:1080"},
		{"vless missing uuid", "vless://@example.com:443?encryption=none"},
		{"vless missing encryption", "vless://uuid@example.com:443"},
		{"vless xhttp missing path", "vless://uuid@example.com:443?encryption=none&type=xhttp&security=reality&pbk=x&sid=y&sni=z&fp=firefox"},
		{"vless reality missing pbk", "vless://uuid@example.com:443?encryption=none&type=tcp&security=reality&sid=y&sni=z&fp=firefox"},
		{"vmess bad base64", "vmess://not-base64!!!"},
		{"vmess missing id", "vmess://" + base64.StdEncoding.EncodeToString([]byte(`{"v":"2","add":"h","port":"443"}`))},
		{"trojan missing password", "trojan://@example.com:443"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateClientLink(tc.link); err == nil {
				t.Fatalf("expected error for %q", tc.link)
			}
		})
	}
}
