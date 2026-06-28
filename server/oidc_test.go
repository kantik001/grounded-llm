package main

import (
	"testing"
	"time"
)

func TestAdminSessionSignVerify(t *testing.T) {
	config = &Config{AdminSecret: "test-secret-key-32bytes-minimum!!"}
	oidcCfg.SessionSecret = ""
	payload := adminSessionPayload{
		Subject: "sub-1",
		Email:   "editor@example.com",
		Name:    "Editor",
		Roles:   []string{RoleKBEditor},
		Exp:     time.Now().Add(time.Hour).Unix(),
		Auth:    adminSessionAuthOIDC,
	}
	token, err := signAdminSession(payload)
	if err != nil {
		t.Fatal(err)
	}
	got, err := verifyAdminSession(token)
	if err != nil {
		t.Fatal(err)
	}
	if got.Email != payload.Email || len(got.Roles) != 1 {
		t.Fatalf("got %+v", got)
	}
}

func TestAdminSessionExpired(t *testing.T) {
	config = &Config{AdminSecret: "test-secret-key-32bytes-minimum!!"}
	payload := adminSessionPayload{
		Subject: "sub",
		Roles:   []string{RoleAdmin},
		Exp:     time.Now().Add(-time.Hour).Unix(),
		Auth:    adminSessionAuthOIDC,
	}
	token, err := signAdminSession(payload)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := verifyAdminSession(token); err == nil {
		t.Fatal("expected expired session error")
	}
}

func TestResolveOIDCRolesFromGroups(t *testing.T) {
	oidcRoleMap = oidcRoleMapping{
		DefaultRoles: []string{RoleKBEditor},
		Claim:        "groups",
		Groups: map[string][]string{
			"grounded-admins": {"admin"},
		},
	}
	roles := resolveOIDCRoles("user@example.com", map[string]any{
		"groups": []any{"grounded-admins"},
	})
	if len(roles) != 1 || roles[0] != RoleAdmin {
		t.Fatalf("got %v", roles)
	}
}

func TestResolveOIDCRolesDefault(t *testing.T) {
	oidcRoleMap = oidcRoleMapping{
		DefaultRoles: []string{RoleKBEditor},
		Groups:       map[string][]string{},
		Emails:       map[string][]string{},
	}
	roles := resolveOIDCRoles("nobody@example.com", map[string]any{})
	if len(roles) != 1 || roles[0] != RoleKBEditor {
		t.Fatalf("got %v", roles)
	}
}

func TestResolveOIDCRolesFromEmail(t *testing.T) {
	oidcRoleMap = oidcRoleMapping{
		DefaultRoles: []string{RoleKBEditor},
		Emails: map[string][]string{
			"admin@company.com": {RoleAdmin},
		},
	}
	roles := resolveOIDCRoles("admin@company.com", map[string]any{})
	if len(roles) != 1 || roles[0] != RoleAdmin {
		t.Fatalf("got %v", roles)
	}
}

func TestAdminAuthEnabledOIDC(t *testing.T) {
	oidcCfg.Enabled = true
	cfg := &Config{}
	if !adminAuthEnabled(cfg) {
		t.Fatal("expected admin enabled when OIDC on")
	}
	oidcCfg.Enabled = false
}
