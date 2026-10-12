package core

import "testing"

func TestQueryXrayStatGRPCInvalidAddr(t *testing.T) {
	if _, err := queryXrayStatGRPC("", "user>>>a>>>traffic>>>uplink"); err == nil {
		t.Fatal("expected error for empty address")
	}
	if _, err := queryXrayStatGRPC("not-a-host-port", "user>>>a>>>traffic>>>uplink"); err == nil {
		t.Fatal("expected error for invalid address")
	}
}
