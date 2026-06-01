package main

import "testing"

func TestLookupAPIKey(t *testing.T) {
	apiKeyRegistry = map[string]string{"secret-key": "demo"}
	if _, ok := lookupAPIKey("secret-key"); !ok {
		t.Fatal("expected key")
	}
	if _, ok := lookupAPIKey("wrong"); ok {
		t.Fatal("unexpected key")
	}
}

func TestAPIKeyActorIDStable(t *testing.T) {
	a := apiKeyActorID("same-key")
	b := apiKeyActorID("same-key")
	if a != b || a >= 0 {
		t.Fatalf("got %d %d", a, b)
	}
}
