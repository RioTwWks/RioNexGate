package api

import (
	"context"
	"net"
	"testing"
)

func TestCheckNodeTCPReachable(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	_, p, _ := net.SplitHostPort(ln.Addr().String())
	var port int
	fmtSscanf := func(s string) int {
		n := 0
		for _, c := range s {
			n = n*10 + int(c-'0')
		}
		return n
	}
	port = fmtSscanf(p)

	blocked := checkNodeTCP(context.Background(), "127.0.0.1", port, false)
	if blocked.Reachable {
		t.Fatal("loopback should be blocked when allowPrivate=false")
	}

	ok := checkNodeTCP(context.Background(), "127.0.0.1", port, true)
	if !ok.Reachable {
		t.Fatalf("expected reachable with allowPrivate: %s", ok.Error)
	}
}
