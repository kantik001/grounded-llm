package admin

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
	CtxKeyAdminRoles = "admin_roles"
	CtxKeyAdminActor = "admin_actor"
	CtxKeyAPIRoles   = "api_roles"
)

var AllRoles = []string{RoleChatOnly, RoleKBEditor, RoleAdmin, RoleAPIManager}

// NormalizeRoles deduplicates and canonicalizes role names.
func NormalizeRoles(in []string) []string {
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

// DefaultAPIKeyRoles is the fallback when an API key has no explicit roles.
func DefaultAPIKeyRoles() []string {
	return []string{RoleChatOnly}
}

// HasAdminRole returns true if actor has admin (superuser) or one of allowed roles.
func HasAdminRole(actorRoles []string, allowed ...string) bool {
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

// CanUseChatAPI reports whether API key roles may call chat endpoints.
func CanUseChatAPI(apiRoles []string) bool {
	return HasAdminRole(apiRoles, RoleChatOnly)
}

// RolesFromContext returns admin roles stored by admin auth middleware.
func RolesFromContext(c *gin.Context) []string {
	if v, ok := c.Get(CtxKeyAdminRoles); ok {
		if roles, ok := v.([]string); ok {
			return roles
		}
	}
	return nil
}

// RequireRoles aborts with 403 unless the actor has one of the allowed roles.
func RequireRoles(allowed ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles := RolesFromContext(c)
		if !HasAdminRole(roles, allowed...) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Forbidden: insufficient role",
			})
			return
		}
		c.Next()
	}
}
