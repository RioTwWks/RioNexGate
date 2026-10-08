package models

import (
	"testing"
	"time"
)

func TestUserAccessAllowed(t *testing.T) {
	u := &User{Active: true, TrafficGB: 1, UsedBytes: 100}
	if !u.AccessAllowed() {
		t.Fatal("expected allowed")
	}
	u.ExpiresAt = time.Now().Add(-time.Hour)
	if u.AccessAllowed() {
		t.Fatal("expired should deny")
	}
	u.ExpiresAt = time.Time{}
	u.UsedBytes = 2 * 1024 * 1024 * 1024
	if u.AccessAllowed() {
		t.Fatal("quota exceeded should deny")
	}
}
