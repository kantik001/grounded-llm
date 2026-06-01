package main

import (
	"github.com/gin-gonic/gin"
)

func registerPublicRoutes(router *gin.Engine) {
	router.GET("/health", handleHealthCheck)
	router.GET("/api/health", handleHealthCheck)
	router.GET("/domains", handleListDomains)
	router.GET("/api/domains", handleListDomains)
	router.GET("/crops", handleListCropsLegacy)
	router.GET("/api/crops", handleListCropsLegacy)
	router.GET("/onboarding", handleOnboarding)
	router.GET("/api/onboarding", handleOnboarding)
	router.GET("/branding", handleBranding)
	router.GET("/api/branding", handleBranding)
}

func mountProtectedAPI(r gin.IRoutes, auth, lim gin.HandlerFunc) {
	deprecated := deprecatedAPIMiddleware()
	r.POST("/chat", auth, lim, deprecated, handleChat)
	r.POST("/session", auth, lim, handleNewSession)
	r.GET("/history", auth, lim, handleHistory)
	r.POST("/message", auth, lim, handleMessage)
	r.POST("/feedback", auth, lim, handleFeedback)
	r.GET("/media/:token", auth, lim, handleMedia)
}

func registerProtectedRoutes(router *gin.Engine, cfg *Config, rl *rateLimiter) {
	auth := telegramAuthMiddleware(cfg)
	lim := rateLimitMiddleware(rl)
	mountProtectedAPI(router.Group(""), auth, lim)
	mountProtectedAPI(router.Group("/api"), auth, lim)
}

func deprecatedAPIMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Deprecation", "true")
		c.Header("Link", "</message>; rel=\"successor-version\"")
		c.Next()
	}
}
