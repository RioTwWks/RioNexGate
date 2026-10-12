package db

import (
	"testing"
	"time"

	"rionexgate/internal/models"
)

func TestCountActiveClientsIncludesTrafficAndDevices(t *testing.T) {
	d, err := Open(t.TempDir() + "/clients.db")
	if err != nil {
		t.Fatal(err)
	}
	if err := d.AutoMigrate(); err != nil {
		t.Fatal(err)
	}

	trafficUser, err := d.CreateUser(CreateUserInput{Email: "traffic@example.com", TrafficGB: 1, ExpireDays: 1})
	if err != nil {
		t.Fatal(err)
	}
	deviceUser, err := d.CreateUser(CreateUserInput{Email: "device@example.com", TrafficGB: 1, ExpireDays: 1})
	if err != nil {
		t.Fatal(err)
	}
	idleUser, err := d.CreateUser(CreateUserInput{Email: "idle@example.com", TrafficGB: 1, ExpireDays: 1})
	if err != nil {
		t.Fatal(err)
	}
	_ = idleUser

	since := time.Now().Add(-24 * time.Hour)
	n, err := d.CountActiveClients(since)
	if err != nil || n != 0 {
		t.Fatalf("want 0 before activity, got %d err=%v", n, err)
	}

	if err := d.TouchUserLastActive(trafficUser.ID); err != nil {
		t.Fatal(err)
	}
	n, err = d.CountActiveClients(since)
	if err != nil || n != 1 {
		t.Fatalf("want 1 after traffic touch, got %d err=%v", n, err)
	}

	dev := models.Device{UserID: deviceUser.ID, Token: "tok-cli", Label: "phone"}
	if err := d.Create(&dev).Error; err != nil {
		t.Fatal(err)
	}
	if err := d.UpsertClientStats(ClientStatsInput{DeviceToken: "tok-cli", SessionID: "s1", BytesIn: 1, BytesOut: 2}); err != nil {
		t.Fatal(err)
	}
	n, err = d.CountActiveClients(since)
	if err != nil || n != 2 {
		t.Fatalf("want 2 after device stats, got %d err=%v", n, err)
	}

	// Same user via both traffic and device must not double-count.
	if err := d.TouchUserLastActive(deviceUser.ID); err != nil {
		t.Fatal(err)
	}
	n, err = d.CountActiveClients(since)
	if err != nil || n != 2 {
		t.Fatalf("want 2 after overlap, got %d err=%v", n, err)
	}
}
