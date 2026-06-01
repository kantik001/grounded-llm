package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// combinedAuthMiddleware: X-API-Key или Telegram initData.
func combinedAuthMiddleware(cfg *Config) gin.HandlerFunc {
	tg := telegramAuthMiddleware(cfg)
	return func(c *gin.Context) {
		key := strings.TrimSpace(c.GetHeader(headerAPIKey))
		if key != "" {
			label, ok := lookupAPIKey(key)
			if !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error":   "Недействительный API key",
				})
				return
			}
			actorID := apiKeyActorID(key)
			c.Set(ctxKeyAPIActorID, actorID)
			c.Set(ctxKeyTelegramUserID, actorID)
			c.Set(ctxKeyTelegramUser, &TelegramUser{ID: actorID, Username: "api:" + label})
			c.Set(ctxKeyAPIKeyLabel, label)
			c.Next()
			return
		}
		tg(c)
	}
}

func ctxActorUser(c *gin.Context) (*TelegramUser, error) {
	return ctxTelegramUser(c)
}

func rateLimitKey(c *gin.Context) string {
	if label, ok := c.Get(ctxKeyAPIKeyLabel); ok {
		if s, ok := label.(string); ok && s != "" {
			return "api:" + s
		}
	}
	if rawID, ok := c.Get(ctxKeyTelegramUserID); ok {
		if id, ok := rawID.(int64); ok && id != 0 {
			return "tg:" + itoa64(id)
		}
	}
	return "anon"
}

func itoa64(v int64) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
