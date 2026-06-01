package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ctxKeyTelegramUserID = "telegram_user_id"
	ctxKeyTelegramUser   = "telegram_user"
	headerTelegramInit   = "X-Telegram-Init-Data"
)

// Разбирает CORS_ALLOWED_ORIGINS в список origin (через запятую).
func parseAllowedOrigins(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		p := strings.TrimSpace(part)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// corsMiddleware разрешает запросы только с перечисленных Origin (безопаснее, чем *).
func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	allowAll := len(allowedOrigins) == 0
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[o] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			if allowAll {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			} else if _, ok := originSet[origin]; ok {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			}
			c.Writer.Header().Set("Vary", "Origin")
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Requested-With, X-Telegram-Init-Data, Authorization, X-API-Key, X-Tenant-ID, X-Locale, X-Request-ID, Accept-Language")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// telegramAuthMiddleware проверяет подпись initData от Telegram Web App.
func telegramAuthMiddleware(cfg *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.TelegramAuthDisabled {
			devID := int64(0)
			if h := strings.TrimSpace(c.GetHeader("X-Dev-User-Id")); h != "" {
				if id, err := strconv.ParseInt(h, 10, 64); err == nil {
					devID = id
				}
			}
			if devID == 0 {
				devID = 1
			}
			c.Set(ctxKeyTelegramUserID, devID)
			c.Next()
			return
		}

		initData := strings.TrimSpace(c.GetHeader(headerTelegramInit))
		if initData == "" {
			// Заголовок Authorization: tma <initData>
			auth := strings.TrimSpace(c.GetHeader("Authorization"))
			if strings.HasPrefix(strings.ToLower(auth), "tma ") {
				initData = strings.TrimSpace(auth[4:])
			}
		}
		if initData == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Telegram authorization required (header X-Telegram-Init-Data). Open the app from the bot.",
			})
			return
		}

		user, err := validateTelegramInitData(initData, cfg.TelegramBotToken, cfg.TelegramInitDataMaxAge)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid Telegram authorization: " + err.Error(),
			})
			return
		}

		c.Set(ctxKeyTelegramUserID, user.ID)
		c.Set(ctxKeyTelegramUser, user)
		c.Next()
	}
}

