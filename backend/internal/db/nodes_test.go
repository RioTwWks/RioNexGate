package db

import (
	"testing"

	"rionexgate/internal/models"
)

func TestNodeCRUDAndChainResolution(t *testing.T) {
	d, err := Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	if err := d.AutoMigrate(); err != nil {
		t.Fatal(err)
	}

	entry, err := d.CreateNode(CreateNodeInput{
		Name: "entry-ru", Address: "entry.ru.example", Port: 443, Role: models.NodeRoleEntry, Priority: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	exit, err := d.CreateNode(CreateNodeInput{
		Name: "exit-eu", Address: "exit.eu.example", Port: 8443, Role: models.NodeRoleExit, Priority: 10,
		Credentials: `{"uuid":"relay"}`,
	})
	if err != nil {
		t.Fatal(err)
	}

	user, err := d.CreateUser(CreateUserInput{Email: "chain@example.com", TrafficGB: 1, ExpireDays: 1})
	if err != nil {
		t.Fatal(err)
	}
	user, err = d.UpdateUser(user.ID, UpdateUserInput{EntryNodeID: &entry.ID, ExitNodeID: &exit.ID})
	if err != nil {
		t.Fatal(err)
	}

	resolvedEntry, err := d.ResolveUserEntryNode(user)
	if err != nil || resolvedEntry.ID != entry.ID {
		t.Fatalf("entry resolve: %+v err=%v", resolvedEntry, err)
	}
	resolvedExit, err := d.ResolveUserExitNode(user)
	if err != nil || resolvedExit.ID != exit.ID {
		t.Fatalf("exit resolve: %+v err=%v", resolvedExit, err)
	}

	if err := d.DeleteNode(exit.ID); err != nil {
		t.Fatal(err)
	}
	user, _ = d.GetUser(user.ID)
	if user.ExitNodeID != nil {
		t.Fatal("exit_node_id should be cleared after node delete")
	}
}
