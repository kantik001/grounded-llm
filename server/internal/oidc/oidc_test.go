package oidc

import (
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/config"
)

func TestAdminSessionSignVerify(t *testing.T) {
	BindConfig(func() *config.Config {
		return &config.Config{AdminSecret: "test-secret-key-32bytes-minimum!!"}
	})
	envCfg.SessionSecret = ""
	payload := SessionPayload{
		Subject: "sub-1",
		Email:   "editor@example.com",
		Name:    "Editor",
		Roles:   []string{defaultEditorRole},
		Exp:     time.Now().Add(time.Hour).Unix(),
		Auth:    SessionAuthOIDC,
	}
	token, err := SignSession(payload)
	if err != nil {
		t.Fatal(err)
	}
	got, err := VerifySession(token)
	if err != nil {
		t.Fatal(err)
	}
	if got.Email != payload.Email || len(got.Roles) != 1 {
		t.Fatalf("got %+v", got)
	}
}

func TestAdminSessionExpired(t *testing.T) {
	BindConfig(func() *config.Config {
		return &config.Config{AdminSecret: "test-secret-key-32bytes-minimum!!"}
	})
	payload := SessionPayload{
		Subject: "sub",
		Roles:   []string{defaultAdminRole},
		Exp:     time.Now().Add(-time.Hour).Unix(),
		Auth:    SessionAuthOIDC,
	}
	token, err := SignSession(payload)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySession(token); err == nil {
		t.Fatal("expected expired session error")
	}
}

func TestResolveOIDCRolesFromGroups(t *testing.T) {
	BindHost(stubHost{})
	roleMap = roleMapping{
		DefaultRoles: []string{defaultEditorRole},
		Claim:        "groups",
		Groups: map[string][]string{
			"grounded-admins": {defaultAdminRole},
		},
	}
	roles := ResolveRoles("user@example.com", map[string]any{
		"groups": []any{"grounded-admins"},
	})
	if len(roles) != 1 || roles[0] != defaultAdminRole {
		t.Fatalf("got %v", roles)
	}
}

func TestResolveOIDCRolesDefault(t *testing.T) {
	BindHost(stubHost{})
	roleMap = roleMapping{
		DefaultRoles: []string{defaultEditorRole},
		Groups:       map[string][]string{},
		Emails:       map[string][]string{},
	}
	roles := ResolveRoles("nobody@example.com", map[string]any{})
	if len(roles) != 1 || roles[0] != defaultEditorRole {
		t.Fatalf("got %v", roles)
	}
}

func TestResolveOIDCRolesFromEmail(t *testing.T) {
	BindHost(stubHost{})
	roleMap = roleMapping{
		DefaultRoles: []string{defaultEditorRole},
		Emails: map[string][]string{
			"admin@company.com": {defaultAdminRole},
		},
	}
	roles := ResolveRoles("admin@company.com", map[string]any{})
	if len(roles) != 1 || roles[0] != defaultAdminRole {
		t.Fatalf("got %v", roles)
	}
}

type stubHost struct{}

func (stubHost) NormalizeRoles(in []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, r := range in {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		if _, ok := seen[r]; ok {
			continue
		}
		seen[r] = struct{}{}
		out = append(out, r)
	}
	return out
}

func (stubHost) RecordAudit(c *gin.Context, opts AuditOpts) {}
func (stubHost) BasicAuthEnabled() bool                     { return false }
