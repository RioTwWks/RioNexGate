package models

import "testing"

func TestNodeCredentialsNormalizeForOutbound(t *testing.T) {
	creds := NodeCredentials{
		UUID:      "relay-uuid",
		PublicKey: "pk",
		ShortID:   "ab12",
	}
	out := creds.NormalizeForOutbound("exit.example.com")
	if out.Security != "reality" {
		t.Fatalf("expected security reality, got %q", out.Security)
	}
	if out.Flow != "xtls-rprx-vision" {
		t.Fatalf("expected vision flow, got %q", out.Flow)
	}
}

func TestNodeCredentialsNormalizePreservesExplicitValues(t *testing.T) {
	creds := NodeCredentials{
		UUID:      "relay-uuid",
		PublicKey: "pk",
		ShortID:   "ab12",
		Security:  "tls",
		Network:   "xhttp",
	}
	out := creds.NormalizeForOutbound("exit.example.com")
	if out.Security != "tls" {
		t.Fatalf("expected tls preserved, got %q", out.Security)
	}
	if out.Flow != "" {
		t.Fatalf("expected empty flow for xhttp, got %q", out.Flow)
	}
}

func TestNodeCredentialsOutboundSNI(t *testing.T) {
	withSNI := NodeCredentials{SNI: "www.cloudflare.com"}
	if withSNI.OutboundSNI("exit.example.com") != "www.cloudflare.com" {
		t.Fatal("expected explicit sni")
	}
	empty := NodeCredentials{}
	if empty.OutboundSNI("exit.example.com") != "exit.example.com" {
		t.Fatal("expected address fallback")
	}
}
