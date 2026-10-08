package netutil

import (
	"context"
	"errors"
	"net"
	"testing"
)

func TestIsBlockedIP(t *testing.T) {
	cases := []struct {
		ip      string
		blocked bool
	}{
		{"127.0.0.1", true},
		{"10.0.0.1", true},
		{"192.168.1.1", true},
		{"172.16.0.1", true},
		{"169.254.1.1", true},
		{"169.254.169.254", true},
		{"0.0.0.0", true},
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"::1", true},
		{"fc00::1", true},
	}
	for _, tc := range cases {
		ip := net.ParseIP(tc.ip)
		if ip == nil {
			t.Fatalf("invalid ip %s", tc.ip)
		}
		got := IsBlockedIP(ip)
		if got != tc.blocked {
			t.Fatalf("%s: blocked=%v want %v", tc.ip, got, tc.blocked)
		}
	}
}

func TestValidateHost(t *testing.T) {
	ctx := context.Background()
	if err := ValidateHost(ctx, "127.0.0.1", false); !errors.Is(err, ErrBlockedHost) {
		t.Fatalf("expected ErrBlockedHost, got %v", err)
	}
	if err := ValidateHost(ctx, "127.0.0.1", true); err != nil {
		t.Fatalf("allowPrivate should accept loopback: %v", err)
	}
	if err := ValidateHost(ctx, "8.8.8.8", false); err != nil {
		t.Fatalf("public IP should be allowed: %v", err)
	}
}
