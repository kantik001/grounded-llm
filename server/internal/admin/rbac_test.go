package admin

import (
	"testing"

	cfgpkg "grounded_llm_server/internal/config"
)

func TestNormalizeRoles(t *testing.T) {
	got := NormalizeRoles([]string{"admin", "kb-editor", "bogus", "chat"})
	if len(got) != 3 {
		t.Fatalf("got %v", got)
	}
	if got[0] != RoleAdmin || got[1] != RoleKBEditor || got[2] != RoleChatOnly {
		t.Fatalf("got %v", got)
	}
}

func TestHasAdminRoleSuperuser(t *testing.T) {
	if !HasAdminRole([]string{RoleAdmin}, RoleKBEditor) {
		t.Fatal("admin should imply kb_editor permission")
	}
}

func TestHasAdminRoleSpecific(t *testing.T) {
	if !HasAdminRole([]string{RoleKBEditor}, RoleKBEditor) {
		t.Fatal("kb_editor should match")
	}
	if HasAdminRole([]string{RoleChatOnly}, RoleKBEditor) {
		t.Fatal("chat_only should not access kb")
	}
	if HasAdminRole([]string{RoleAPIManager}, RoleKBEditor) {
		t.Fatal("api_manager should not access kb")
	}
}

func TestCanUseChatAPI(t *testing.T) {
	if !CanUseChatAPI([]string{RoleChatOnly}) {
		t.Fatal("chat_only should access chat API")
	}
	if CanUseChatAPI([]string{RoleAPIManager}) {
		t.Fatal("api_manager-only key should not access chat")
	}
}

func TestAuthenticateAdminLegacy(t *testing.T) {
	BindConfig(func() *cfgpkg.Config {
		return &cfgpkg.Config{AdminUser: "admin", AdminPassword: "secret"}
	})
	userRegistry = nil
	roles, ok := authenticateUser("admin", "secret")
	if !ok || len(roles) != 1 || roles[0] != RoleAdmin {
		t.Fatalf("got ok=%v roles=%v", ok, roles)
	}
	if _, ok := authenticateUser("admin", "wrong"); ok {
		t.Fatal("expected auth failure")
	}
}

func TestAuthenticateAdminUsersFile(t *testing.T) {
	BindConfig(func() *cfgpkg.Config {
		return &cfgpkg.Config{AdminUser: "legacy", AdminPassword: "x"}
	})
	userRegistry = map[string]userRecord{
		"editor": {
			Username: "editor",
			Password: "pass",
			Roles:    []string{RoleKBEditor},
		},
	}
	roles, ok := authenticateUser("editor", "pass")
	if !ok || len(roles) != 1 || roles[0] != RoleKBEditor {
		t.Fatalf("got ok=%v roles=%v", ok, roles)
	}
}
