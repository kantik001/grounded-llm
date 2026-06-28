package main

import "testing"

func TestNormalizeRoles(t *testing.T) {
	got := normalizeRoles([]string{"admin", "kb-editor", "bogus", "chat"})
	if len(got) != 3 {
		t.Fatalf("got %v", got)
	}
	if got[0] != RoleAdmin || got[1] != RoleKBEditor || got[2] != RoleChatOnly {
		t.Fatalf("got %v", got)
	}
}

func TestHasAdminRoleSuperuser(t *testing.T) {
	if !hasAdminRole([]string{RoleAdmin}, RoleKBEditor) {
		t.Fatal("admin should imply kb_editor permission")
	}
}

func TestHasAdminRoleSpecific(t *testing.T) {
	if !hasAdminRole([]string{RoleKBEditor}, RoleKBEditor) {
		t.Fatal("kb_editor should match")
	}
	if hasAdminRole([]string{RoleChatOnly}, RoleKBEditor) {
		t.Fatal("chat_only should not access kb")
	}
	if hasAdminRole([]string{RoleAPIManager}, RoleKBEditor) {
		t.Fatal("api_manager should not access kb")
	}
}

func TestCanUseChatAPI(t *testing.T) {
	if !canUseChatAPI([]string{RoleChatOnly}) {
		t.Fatal("chat_only should access chat API")
	}
	if canUseChatAPI([]string{RoleAPIManager}) {
		t.Fatal("api_manager-only key should not access chat")
	}
}

func TestAuthenticateAdminLegacy(t *testing.T) {
	config = &Config{AdminUser: "admin", AdminPassword: "secret"}
	adminUserRegistry = nil
	roles, ok := authenticateAdminUser("admin", "secret")
	if !ok || len(roles) != 1 || roles[0] != RoleAdmin {
		t.Fatalf("got ok=%v roles=%v", ok, roles)
	}
	if _, ok := authenticateAdminUser("admin", "wrong"); ok {
		t.Fatal("expected auth failure")
	}
}

func TestAuthenticateAdminUsersFile(t *testing.T) {
	config = &Config{AdminUser: "legacy", AdminPassword: "x"}
	adminUserRegistry = map[string]adminUserRecord{
		"editor": {
			Username: "editor",
			Password: "pass",
			Roles:    []string{RoleKBEditor},
		},
	}
	roles, ok := authenticateAdminUser("editor", "pass")
	if !ok || len(roles) != 1 || roles[0] != RoleKBEditor {
		t.Fatalf("got ok=%v roles=%v", ok, roles)
	}
}
