package main

import "testing"

func TestCoalesceDomainID(t *testing.T) {
	if got := coalesceDomainID("", "  default ", ""); got != "default" {
		t.Fatalf("got %q", got)
	}
	if got := coalesceDomainID("hr", "legal"); got != "hr" {
		t.Fatalf("got %q", got)
	}
	if got := coalesceDomainID("", ""); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}
