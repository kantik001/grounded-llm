package main

import (
	"github.com/gin-gonic/gin"
)

func registerPublicRoutes(router *gin.Engine) {
	router.GET("/health", handleHealthCheck)
	router.GET("/api/health", handleHealthCheck)
	router.GET("/metrics", handleMetrics)
	router.GET("/api/metrics", handleMetrics)
	router.GET("/domains", handleListDomains)
	router.GET("/api/domains", handleListDomains)
	router.GET("/onboarding", handleOnboarding)
	router.GET("/api/onboarding", handleOnboarding)
	router.GET("/branding", handleBranding)
	router.GET("/api/branding", handleBranding)
}

func mountProtectedAPI(r gin.IRoutes) {
	r.POST("/session", handleNewSession)
	r.GET("/history", handleHistory)
	r.POST("/message", handleMessage)
	r.POST("/feedback", handleFeedback)
	r.GET("/media/:token", handleMedia)
}

func registerProtectedRoutes(router *gin.Engine, cfg *Config, rl *rateLimiter) {
	auth := combinedAuthMiddleware(cfg)
	lim := rateLimitMiddleware(rl)
	tenant := tenantMiddleware(cfg)

	legacy := router.Group("")
	legacy.Use(tenant)
	legacy.Use(auth)
	legacy.Use(lim)
	mountProtectedAPI(legacy)

	api := router.Group("/api")
	api.Use(tenant)
	api.Use(auth)
	api.Use(lim)
	mountProtectedAPI(api)

	v1 := router.Group("/api/v1")
	v1.Use(tenant)
	v1.Use(auth)
	v1.Use(lim)
	mountProtectedAPI(v1)
	v1.GET("/openapi.json", handleOpenAPI)
}
