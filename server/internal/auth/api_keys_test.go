package auth

import (
	"testing"

	"grounded_llm_server/internal/config"
)

func TestLookupAPIKey(t *testing.T) {
	Registry = map[string]KeyRecord{
		HashKey("secret-key"): {Label: "demo", Roles: []string{RoleChatOnly}},
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
		HashKey("k"): {Label: "x", Roles: nil},
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

func TestLoadAPIKeysHashesPlaintext(t *testing.T) {
	t.Setenv("API_KEYS_FILE", "")
	t.Setenv("API_KEYS", "plain-key:integrator:acme")
	LoadAPIKeys(&config.Config{})

	if _, exists := Registry["plain-key"]; exists {
		t.Fatal("plaintext key must not be a registry key")
	}
	rec, ok := Lookup("plain-key")
	if !ok {
		t.Fatal("expected key via hashed lookup")
	}
	if rec.Label != "integrator" || rec.Tenant != "acme" {
		t.Fatalf("rec=%+v", rec)
	}
}

func TestLoadAPIKeysAcceptsPreHashed(t *testing.T) {
	t.Setenv("API_KEYS_FILE", "")
	t.Setenv("API_KEYS", "sha256:"+HashKey("raw-secret")+":ci:*")
	LoadAPIKeys(&config.Config{})

	rec, ok := Lookup("raw-secret")
	if !ok {
		t.Fatal("expected pre-hashed key to match raw secret")
	}
	if rec.Label != "ci" || rec.Tenant != "*" {
		t.Fatalf("rec=%+v", rec)
	}
}

func TestLoadAPIKeysLegacyFormatStillWorks(t *testing.T) {
	t.Setenv("API_KEYS_FILE", "")
	t.Setenv("API_KEYS", "old-key:legacy-label")
	LoadAPIKeys(&config.Config{})

	rec, ok := Lookup("old-key")
	if !ok || rec.Label != "legacy-label" || rec.Tenant != "" {
		t.Fatalf("rec=%+v ok=%v", rec, ok)
	}
}
