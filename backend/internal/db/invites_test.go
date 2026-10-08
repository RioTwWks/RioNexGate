package db

import (
	"path/filepath"
	"testing"
	"time"
)

func TestInviteCreateConsumeRevoke(t *testing.T) {
	if testing.Short() {
		t.Skip("sqlite")
	}
	database, err := Open(filepath.Join(t.TempDir(), "inv.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(); err != nil {
		t.Fatal(err)
	}
	user, err := database.CreateUser(CreateUserInput{Email: "inv@example.com", TrafficGB: 1, ExpireDays: 7})
	if err != nil {
		t.Fatal(err)
	}

	inv, err := database.CreateInvite(CreateInviteInput{
		UserID:    user.ID,
		Label:     "phone",
		MaxUses:   1,
		ExpiresIn: time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !inv.Usable() {
		t.Fatal("expected usable invite")
	}

	got, err := database.ConsumeInvite(inv.Token)
	if err != nil {
		t.Fatal(err)
	}
	if got.UsedCount != 1 {
		t.Fatalf("used_count=%d", got.UsedCount)
	}
	if _, err := database.ConsumeInvite(inv.Token); err == nil {
		t.Fatal("expected second consume to fail")
	}

	inv2, err := database.CreateInvite(CreateInviteInput{UserID: user.ID, MaxUses: 2})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.RevokeInvite(user.ID, inv2.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ConsumeInvite(inv2.Token); err == nil {
		t.Fatal("expected revoked invite to fail")
	}
}
