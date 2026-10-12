package db

import (
	"testing"
	"time"

	"rionexgate/internal/models"
)

func TestCountActiveDevicesIncludesStatsReports(t *testing.T) {
	d, err := Open(t.TempDir() + "/devcount.db")
	if err != nil {
		t.Fatal(err)
	}
	if err := d.AutoMigrate(); err != nil {
		t.Fatal(err)
	}
	user, err := d.CreateUser(CreateUserInput{Email: "d@example.com", TrafficGB: 1, ExpireDays: 1})
	if err != nil {
		t.Fatal(err)
	}
	dev := models.Device{UserID: user.ID, Token: "tok-active", Label: "phone"}
	if err := d.Create(&dev).Error; err != nil {
		t.Fatal(err)
	}
	since := time.Now().Add(-24 * time.Hour)
	n, err := d.CountActiveDevices(since)
	if err != nil || n != 0 {
		t.Fatalf("want 0 before activity, got %d err=%v", n, err)
	}
	if err := d.UpsertClientStats(ClientStatsInput{DeviceToken: "tok-active", SessionID: "s1", BytesIn: 10, BytesOut: 20}); err != nil {
		t.Fatal(err)
	}
	n, err = d.CountActiveDevices(since)
	if err != nil || n != 1 {
		t.Fatalf("want 1 after stats report, got %d err=%v", n, err)
	}
}
