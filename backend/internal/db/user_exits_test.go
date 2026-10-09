package db

import (
	"testing"

	"rionexgate/internal/models"
)

func TestSetUserExitsMultiAndPreserveUUID(t *testing.T) {
	d, err := Open(t.TempDir() + "/user_exits.db")
	if err != nil {
		t.Fatal(err)
	}
	if err := d.AutoMigrate(); err != nil {
		t.Fatal(err)
	}

	nl, err := d.CreateNode(CreateNodeInput{
		Name: "nl", Address: "nl.example", Port: 8443, Role: models.NodeRoleExit, Region: "NL", Priority: 10,
		Credentials: `{"uuid":"relay-nl"}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	de, err := d.CreateNode(CreateNodeInput{
		Name: "de", Address: "de.example", Port: 8443, Role: models.NodeRoleExit, Region: "DE", Priority: 20,
		Credentials: `{"uuid":"relay-de"}`,
	})
	if err != nil {
		t.Fatal(err)
	}

	user, err := d.CreateUser(CreateUserInput{Email: "multi@example.com", TrafficGB: 10, ExpireDays: 30})
	if err != nil {
		t.Fatal(err)
	}

	user, err = d.SetUserExits(user.ID, []uint{nl.ID, de.ID})
	if err != nil {
		t.Fatal(err)
	}
	if user.ExitNodeID == nil || *user.ExitNodeID != nl.ID {
		t.Fatalf("primary exit_node_id want %d, got %+v", nl.ID, user.ExitNodeID)
	}

	rows, err := d.ListUserExits(user.ID)
	if err != nil || len(rows) != 2 {
		t.Fatalf("want 2 exits, got %d err=%v", len(rows), err)
	}
	if rows[0].UUID != user.UUID || rows[0].Email != user.Email {
		t.Fatalf("primary must keep panel identity: %+v", rows[0])
	}
	if rows[1].UUID == user.UUID || rows[1].Email == user.Email {
		t.Fatalf("secondary must have distinct identity: %+v", rows[1])
	}
	secondaryUUID := rows[1].UUID
	secondaryEmail := rows[1].Email

	// Re-save same exits (order swap: DE first becomes primary)
	user, err = d.SetUserExits(user.ID, []uint{de.ID, nl.ID})
	if err != nil {
		t.Fatal(err)
	}
	rows, err = d.ListUserExits(user.ID)
	if err != nil || len(rows) != 2 {
		t.Fatalf("want 2 exits after reorder, got %d err=%v", len(rows), err)
	}
	if rows[0].NodeID != de.ID || rows[0].UUID != user.UUID {
		t.Fatalf("new primary should be DE with panel UUID: %+v", rows[0])
	}
	// NL is now secondary — if it was previously primary (panel UUID), it gets a new UUID.
	// DE was secondary and should keep its previous secondary UUID when it becomes... wait,
	// DE is now primary so it gets panel UUID. NL as secondary: previously had panel UUID,
	// so preservation only applies when prev.UUID != user.UUID. NL will get a fresh UUID.
	_ = secondaryUUID
	_ = secondaryEmail

	// Bind NL+DE again with NL primary; secondary DE should preserve its UUID from when it was secondary before reorder...
	// After reorder DE is primary (panel UUID). Set back NL, DE:
	user, err = d.SetUserExits(user.ID, []uint{nl.ID, de.ID})
	if err != nil {
		t.Fatal(err)
	}
	rows, _ = d.ListUserExits(user.ID)
	if len(rows) != 2 {
		t.Fatalf("want 2, got %d", len(rows))
	}
	deUUID := rows[1].UUID
	deEmail := rows[1].Email

	user, err = d.SetUserExits(user.ID, []uint{nl.ID, de.ID})
	if err != nil {
		t.Fatal(err)
	}
	rows, _ = d.ListUserExits(user.ID)
	if rows[1].UUID != deUUID || rows[1].Email != deEmail {
		t.Fatalf("secondary UUID/email must be stable on re-save: got %+v want %s / %s", rows[1], deUUID, deEmail)
	}

	assignments, err := d.ListUserExitAssignments(user.ID)
	if err != nil || len(assignments) != 2 {
		t.Fatalf("assignments: %d err=%v", len(assignments), err)
	}
	if assignments[0].Node.Region != "NL" || assignments[1].Node.Region != "DE" {
		t.Fatalf("unexpected regions: %+v", assignments)
	}

	emails, err := d.ListUserExitEmails(user.ID)
	if err != nil || len(emails) != 2 {
		t.Fatalf("emails: %v err=%v", emails, err)
	}
}
