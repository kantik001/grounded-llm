package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Built-in RBAC roles (Phase B).
const (
	RoleChatOnly   = "chat_only"
	RoleKBEditor   = "kb_editor"
	RoleAdmin      = "admin"
	RoleAPIManager = "api_manager"
)

const (
	ctxKeyAdminRoles = "admin_roles"
	ctxKeyAPIRoles   = "api_roles"
)

var allRoles = []string{RoleChatOnly, RoleKBEditor, RoleAdmin, RoleAPIManager}

func normalizeRoles(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(in))
	var out []string
	for _, raw := range in {
		r := normalizeRoleName(raw)
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

func normalizeRoleName(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case RoleChatOnly, "chat", "chat-only":
		return RoleChatOnly
	case RoleKBEditor, "kb", "editor", "kb-editor":
		return RoleKBEditor
	case RoleAdmin:
		return RoleAdmin
	case RoleAPIManager, "api", "api-manager":
		return RoleAPIManager
	default:
		return ""
	}
}

func defaultAPIKeyRoles() []string {
	return []string{RoleChatOnly}
}

// hasAdminRole returns true if actor has admin (superuser) or one of allowed roles.
func hasAdminRole(actorRoles []string, allowed ...string) bool {
	if len(actorRoles) == 0 {
		return false
	}
	for _, r := range actorRoles {
		if r == RoleAdmin {
			return true
		}
		for _, a := range allowed {
			if r == a {
				return true
			}
		}
	}
	return false
}

func canUseChatAPI(apiRoles []string) bool {
	return hasAdminRole(apiRoles, RoleChatOnly)
}

func adminRolesFromContext(c *gin.Context) []string {
	if v, ok := c.Get(ctxKeyAdminRoles); ok {
		if roles, ok := v.([]string); ok {
			return roles
		}
	}
	return nil
}

func apiRolesFromContext(c *gin.Context) []string {
	if v, ok := c.Get(ctxKeyAPIRoles); ok {
		if roles, ok := v.([]string); ok {
			return roles
		}
	}
	return nil
}

func requireAdminRoles(allowed ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles := adminRolesFromContext(c)
		if !hasAdminRole(roles, allowed...) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Forbidden: insufficient role",
			})
			return
		}
		c.Next()
	}
}
