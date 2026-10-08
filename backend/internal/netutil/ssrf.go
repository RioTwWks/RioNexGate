package netutil

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

// ErrBlockedHost indicates the target resolves to a disallowed address class.
var ErrBlockedHost = errors.New("host resolves to a blocked address (loopback/private/link-local/metadata)")

// IsBlockedIP reports whether ip must not be dialed when private probes are disabled.
func IsBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() || ip.IsMulticast() || ip.IsInterfaceLocalMulticast() {
		return true
	}
	if ip.Equal(net.ParseIP("169.254.169.254")) {
		return true
	}
	if ip4 := ip.To4(); ip4 == nil && len(ip) == net.IPv6len {
		if (ip[0] & 0xfe) == 0xfc {
			return true
		}
	}
	return false
}

func normalizeHost(host string) string {
	host = strings.TrimSpace(host)
	host = strings.TrimPrefix(host, "[")
	host = strings.TrimSuffix(host, "]")
	return host
}

// ResolveAllowedIPs returns dialable IPs for host. When allowPrivate is false,
// blocked address classes are filtered out; if none remain, ErrBlockedHost is returned.
func ResolveAllowedIPs(ctx context.Context, host string, allowPrivate bool) ([]net.IP, error) {
	host = normalizeHost(host)
	if host == "" {
		return nil, fmt.Errorf("empty host")
	}

	var ips []net.IP
	if ip := net.ParseIP(host); ip != nil {
		ips = []net.IP{ip}
	} else {
		addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, fmt.Errorf("resolve host: %w", err)
		}
		for _, a := range addrs {
			ips = append(ips, a.IP)
		}
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("host resolved to no addresses")
	}
	if allowPrivate {
		return ips, nil
	}
	var allowed []net.IP
	for _, ip := range ips {
		if !IsBlockedIP(ip) {
			allowed = append(allowed, ip)
		}
	}
	if len(allowed) == 0 {
		return nil, ErrBlockedHost
	}
	return allowed, nil
}

// ValidateHost resolves host and rejects blocked address classes unless allowPrivate.
func ValidateHost(ctx context.Context, host string, allowPrivate bool) error {
	_, err := ResolveAllowedIPs(ctx, host, allowPrivate)
	return err
}

// DialTCP validates the host, then dials a concrete allowed IP (avoids DNS rebinding).
func DialTCP(ctx context.Context, host string, port int, allowPrivate bool, timeout time.Duration) (net.Conn, error) {
	ips, err := ResolveAllowedIPs(ctx, host, allowPrivate)
	if err != nil {
		return nil, err
	}
	dialer := &net.Dialer{Timeout: timeout}
	var lastErr error
	for _, ip := range ips {
		addr := net.JoinHostPort(ip.String(), fmt.Sprintf("%d", port))
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no addresses to dial")
	}
	return nil, lastErr
}
