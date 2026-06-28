package main

import "testing"

func TestIsReindexTerminal(t *testing.T) {
	if !isReindexTerminal(reindexStatusSucceeded) {
		t.Fatal("succeeded should be terminal")
	}
	if !isReindexTerminal(reindexStatusFailed) {
		t.Fatal("failed should be terminal")
	}
	if isReindexTerminal(reindexStatusRunning) {
		t.Fatal("running should not be terminal")
	}
}

func TestReindexStatusLabel(t *testing.T) {
	cases := map[string]string{
		reindexStatusPending:   "queued",
		reindexStatusRunning:   "running",
		reindexStatusSucceeded: "succeeded",
		reindexStatusFailed:    "failed",
	}
	for in, want := range cases {
		if got := reindexStatusLabel(in); got != want {
			t.Fatalf("%q: got %q want %q", in, got, want)
		}
	}
}

func TestReindexAcceptedMessage(t *testing.T) {
	if reindexAcceptedMessage(false) == "" {
		t.Fatal("expected message for new job")
	}
	if reindexAcceptedMessage(true) == reindexAcceptedMessage(false) {
		t.Fatal("already running message should differ")
	}
}
