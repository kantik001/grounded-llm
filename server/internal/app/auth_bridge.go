package app

import (
	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/auth"
)

type apiKeyRecord = auth.KeyRecord

// apiKeyRegistry mirrors auth.Registry after each load (admin RBAC iterates it).
var apiKeyRegistry map[string]apiKeyRecord

func loadAPIKeys(cfg *Config) {
	auth.LoadAPIKeys(cfg)
	apiKeyRegistry = auth.Registry
}

func combinedAuthMiddleware(cfg *Config) gin.HandlerFunc {
	return auth.CombinedMiddleware(cfg)
}

func ctxActorUser(c *gin.Context) (*TelegramUser, error) {
	return auth.CtxActorUser(c)
}

func rateLimitKey(c *gin.Context) string {
	return auth.RateLimitKey(c)
}
