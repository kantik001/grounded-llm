package main

import "testing"

func TestLookupAPIKey(t *testing.T) {
	apiKeyRegistry = map[string]apiKeyRecord{
		"secret-key": {Label: "demo", Roles: []string{RoleChatOnly}},
	}
	rec, ok := lookupAPIKey("secret-key")
	if !ok || rec.Label != "demo" {
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

func TestAPIKeyDefaultRoles(t *testing.T) {
	apiKeyRegistry = map[string]apiKeyRecord{
		"k": {Label: "x", Roles: nil},
	}
	rec, ok := lookupAPIKey("k")
	if !ok {
		t.Fatal("expected key")
	}
	roles := rec.Roles
	if len(roles) == 0 {
		roles = defaultAPIKeyRoles()
	}
	if len(roles) != 1 || roles[0] != RoleChatOnly {
		t.Fatalf("roles=%v", roles)
	}
}
