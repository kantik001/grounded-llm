package httpapi

import "github.com/gin-gonic/gin"

// PublicHandlers are unauthenticated catalog/health endpoints.
type PublicHandlers struct {
	Health     gin.HandlerFunc
	Ready      gin.HandlerFunc
	Metrics    gin.HandlerFunc
	Domains    gin.HandlerFunc
	Onboarding gin.HandlerFunc
	Branding   gin.HandlerFunc
}

// ProtectedHandlers are chat API endpoints behind auth.
type ProtectedHandlers struct {
	Session  gin.HandlerFunc
	History  gin.HandlerFunc
	Message  gin.HandlerFunc
	Feedback gin.HandlerFunc
	Media    gin.HandlerFunc
	OpenAPI  gin.HandlerFunc
}

// Stack is auth/tenant/rate-limit middleware for protected groups.
type Stack struct {
	Tenant    gin.HandlerFunc
	Auth      gin.HandlerFunc
	RateLimit gin.HandlerFunc
}

// RegisterPublic mounts health, metrics, and public catalog routes.
func RegisterPublic(router *gin.Engine, h PublicHandlers) {
	router.GET("/health", h.Health)
	router.GET("/api/health", h.Health)
	router.GET("/ready", h.Ready)
	router.GET("/api/ready", h.Ready)
	router.GET("/metrics", h.Metrics)
	router.GET("/api/metrics", h.Metrics)
	router.GET("/domains", h.Domains)
	router.GET("/api/domains", h.Domains)
	router.GET("/onboarding", h.Onboarding)
	router.GET("/api/onboarding", h.Onboarding)
	router.GET("/branding", h.Branding)
	router.GET("/api/branding", h.Branding)
}

// MountProtected mounts session/history/message/feedback/media on a group.
func MountProtected(r gin.IRoutes, h ProtectedHandlers) {
	r.POST("/session", h.Session)
	r.GET("/history", h.History)
	r.POST("/message", h.Message)
	r.POST("/feedback", h.Feedback)
	r.GET("/media/:token", h.Media)
}

// RegisterProtected mounts legacy + /api + /api/v1 chat routes with the given stack.
func RegisterProtected(router *gin.Engine, stack Stack, h ProtectedHandlers) {
	legacy := router.Group("")
	legacy.Use(stack.Tenant, stack.Auth, stack.RateLimit)
	MountProtected(legacy, h)

	api := router.Group("/api")
	api.Use(stack.Tenant, stack.Auth, stack.RateLimit)
	MountProtected(api, h)

	v1 := router.Group("/api/v1")
	v1.Use(stack.Tenant, stack.Auth, stack.RateLimit)
	MountProtected(v1, h)
	openapi := h.OpenAPI
	if openapi == nil {
		openapi = OpenAPI
	}
	v1.GET("/openapi.json", openapi)
}
