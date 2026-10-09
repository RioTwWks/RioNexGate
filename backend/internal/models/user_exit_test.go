package models

import "testing"

func TestExitSlugAndRouteEmail(t *testing.T) {
	nl := Node{Name: "Amsterdam-1", Region: "NL"}
	if got := ExitSlug(nl); got != "nl" {
		t.Fatalf("ExitSlug: got %q", got)
	}
	if got := ExitDisplayLabel(nl); got != "NL" {
		t.Fatalf("ExitDisplayLabel: got %q", got)
	}
	if got := ExitRouteEmail("u@x", "de", true); got != "u@x" {
		t.Fatalf("primary email: got %q", got)
	}
	if got := ExitRouteEmail("u@x", "de", false); got != "u@x/de" {
		t.Fatalf("secondary email: got %q", got)
	}
}
