package app

import (
	"time"

	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/auth"
)

const (
	ctxKeyAPIKeyLabel    = auth.CtxKeyAPIKeyLabel
	ctxKeyAPIActorID     = auth.CtxKeyAPIActorID
	ctxKeyTelegramUserID = auth.CtxKeyTelegramUserID
	ctxKeyTelegramUser   = auth.CtxKeyTelegramUser
	headerAPIKey         = auth.HeaderAPIKey
	headerTelegramInit   = auth.HeaderTelegramInit
)

type apiKeyRecord = auth.KeyRecord

// apiKeyRegistry mirrors auth.Registry after each load (admin RBAC iterates it).
var apiKeyRegistry map[string]apiKeyRecord

func loadAPIKeys(cfg *Config) {
	auth.LoadAPIKeys(cfg)
	apiKeyRegistry = auth.Registry
}

func lookupAPIKey(key string) (apiKeyRecord, bool) {
	return auth.Lookup(key)
}

func apiKeyActorID(key string) int64 {
	return auth.ActorID(key)
}

func validateTelegramInitData(initData, botToken string, maxAge time.Duration) (*TelegramUser, error) {
	return auth.ValidateTelegramInitData(initData, botToken, maxAge)
}

func telegramAuthMiddleware(cfg *Config) gin.HandlerFunc {
	return auth.TelegramMiddleware(cfg)
}

func combinedAuthMiddleware(cfg *Config) gin.HandlerFunc {
	return auth.CombinedMiddleware(cfg)
}

func ctxTelegramUser(c *gin.Context) (*TelegramUser, error) {
	return auth.CtxTelegramUser(c)
}

func ctxActorUser(c *gin.Context) (*TelegramUser, error) {
	return auth.CtxActorUser(c)
}

func rateLimitKey(c *gin.Context) string {
	return auth.RateLimitKey(c)
}
