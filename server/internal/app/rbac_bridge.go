package app

import (
	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/admin"
	"grounded_llm_server/internal/auth"
)

const (
	RoleChatOnly   = admin.RoleChatOnly
	RoleKBEditor   = admin.RoleKBEditor
	RoleAdmin      = admin.RoleAdmin
	RoleAPIManager = admin.RoleAPIManager
)

const (
	ctxKeyAdminRoles = admin.CtxKeyAdminRoles
	ctxKeyAPIRoles   = auth.CtxKeyAPIRoles
)

func normalizeRoles(in []string) []string {
	return admin.NormalizeRoles(in)
}

func defaultAPIKeyRoles() []string {
	return admin.DefaultAPIKeyRoles()
}

func canUseChatAPI(apiRoles []string) bool {
	return admin.CanUseChatAPI(apiRoles)
}

func hasAdminRole(actorRoles []string, allowed ...string) bool {
	return admin.HasAdminRole(actorRoles, allowed...)
}

func adminRolesFromContext(c *gin.Context) []string {
	return admin.RolesFromContext(c)
}

func requireAdminRoles(allowed ...string) gin.HandlerFunc {
	return admin.RequireRoles(allowed...)
}
