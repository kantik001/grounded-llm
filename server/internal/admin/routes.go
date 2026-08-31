package admin

import (
	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/oidc"
)

// RegisterRoutes mounts /admin and /api/admin route groups.
func RegisterRoutes(router *gin.Engine) {
	oidc.RegisterAuthRoutes(router.Group("/admin/auth"))
	oidc.RegisterAuthRoutes(router.Group("/api/admin/auth"))
	auth := AuthMiddleware()
	registerRouteGroup(router.Group("/admin"), auth)
	registerRouteGroup(router.Group("/api/admin"), auth)
}

func registerRouteGroup(g *gin.RouterGroup, auth gin.HandlerFunc) {
	g.Use(auth)

	g.GET("/status", handleStatus)

	kb := g.Group("")
	kb.Use(RequireRoles(RoleKBEditor))
	kb.GET("/articles", handleListArticles)
	kb.DELETE("/articles", handleDeleteArticle)
	kb.POST("/upload", handleUpload)
	kb.POST("/reindex", handleReindex)
	kb.GET("/reindex/status", handleReindexStatus)
	kb.POST("/ingest", handleIngest)
	kb.GET("/ingest/status", handleIngestStatus)
	kb.GET("/quotas", handleQuotas)

	adminOnly := g.Group("")
	adminOnly.Use(RequireRoles(RoleAdmin))
	adminOnly.GET("/feedback", handleFeedbackSummary)
	adminOnly.GET("/analytics", handleAnalytics)
	adminOnly.GET("/audit-log", handleAuditLog)
	adminOnly.DELETE("/tenants/:tenant_id", handlePurgeTenant)

	apiMgr := g.Group("")
	apiMgr.Use(RequireRoles(RoleAPIManager))
	apiMgr.GET("/api-keys", handleAPIKeys)
}
