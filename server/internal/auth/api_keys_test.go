package auth

import "testing"

func TestLookupAPIKey(t *testing.T) {
	Registry = map[string]KeyRecord{
		"secret-key": {Label: "demo", Roles: []string{RoleChatOnly}},
	}
	rec, ok := Lookup("secret-key")
	if !ok || rec.Label != "demo" {
		t.Fatal("expected key")
	}
	if _, ok := Lookup("wrong"); ok {
		t.Fatal("unexpected key")
	}
}

func TestAPIKeyActorIDStable(t *testing.T) {
	a := ActorID("same-key")
	b := ActorID("same-key")
	if a != b || a >= 0 {
		t.Fatalf("got %d %d", a, b)
	}
}

func TestAPIKeyDefaultRoles(t *testing.T) {
	Registry = map[string]KeyRecord{
		"k": {Label: "x", Roles: nil},
	}
	rec, ok := Lookup("k")
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
