package auth

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/config"
	"grounded_llm_server/internal/store"
	"grounded_llm_server/internal/tenant"
)

const (
	CtxKeyTelegramUserID = "telegram_user_id"
	CtxKeyTelegramUser   = "telegram_user"
	CtxKeyAPIRoles       = "api_roles"
	HeaderTelegramInit   = "X-Telegram-Init-Data"
)

// TelegramMiddleware validates Telegram Web App initData (or dev bypass when disabled).
func TelegramMiddleware(cfg *config.Config) gin.HandlerFunc {
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
			c.Set(CtxKeyTelegramUserID, devID)
			if err := tenant.BindTelegramMembership(c, devID); err != nil {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"success": false,
					"error":   "Forbidden: " + err.Error(),
				})
				return
			}
			c.Next()
			return
		}

		initData := strings.TrimSpace(c.GetHeader(HeaderTelegramInit))
		if initData == "" {
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

		user, err := ValidateTelegramInitData(initData, cfg.TelegramBotToken, cfg.TelegramInitDataMaxAge)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid Telegram authorization: " + err.Error(),
			})
			return
		}

		c.Set(CtxKeyTelegramUserID, user.ID)
		c.Set(CtxKeyTelegramUser, user)
		if err := tenant.BindTelegramMembership(c, user.ID); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Forbidden: " + err.Error(),
			})
			return
		}
		c.Next()
		return
	}
}

// CombinedMiddleware accepts X-API-Key or Telegram initData.
func CombinedMiddleware(cfg *config.Config) gin.HandlerFunc {
	tg := TelegramMiddleware(cfg)
	return func(c *gin.Context) {
		key := strings.TrimSpace(c.GetHeader(HeaderAPIKey))
		if key != "" {
			rec, ok := Lookup(key)
			if !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error":   "Invalid API key",
				})
				return
			}
			roles := rec.Roles
			if len(roles) == 0 {
				roles = defaultAPIKeyRoles()
			}
			if !CanUseChatAPI(roles) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"success": false,
					"error":   "Forbidden: API key cannot access chat",
				})
				return
			}
			// Bind the request to the tenant this key belongs to. Keys without an
			// explicit tenant are scoped to the default tenant; "*" opts out.
			boundTenant := strings.TrimSpace(rec.Tenant)
			if boundTenant == "" {
				boundTenant = cfg.DefaultTenantID
			}
			if err := tenant.BindCredentialTenant(c, boundTenant); err != nil {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"success": false,
					"error":   "Forbidden: API key is not authorized for the requested tenant",
				})
				return
			}
			label := rec.Label
			if label == "" {
				label = "api"
			}
			actorID := ActorID(key)
			c.Set(CtxKeyAPIActorID, actorID)
			c.Set(CtxKeyTelegramUserID, actorID)
			c.Set(CtxKeyTelegramUser, &store.TelegramUser{ID: actorID, Username: "api:" + label})
			c.Set(CtxKeyAPIKeyLabel, label)
			c.Set(CtxKeyAPIRoles, roles)
			c.Next()
			return
		}
		tg(c)
	}
}

// CtxTelegramUser reads TelegramUser from Gin context after auth middleware.
func CtxTelegramUser(c *gin.Context) (*store.TelegramUser, error) {
	if raw, ok := c.Get(CtxKeyTelegramUser); ok {
		if u, ok := raw.(*store.TelegramUser); ok && u != nil {
			return u, nil
		}
	}
	if raw, ok := c.Get(CtxKeyTelegramUserID); ok {
		if id, ok := raw.(int64); ok && id != 0 {
			return &store.TelegramUser{ID: id}, nil
		}
	}
	return nil, fmt.Errorf("telegram user not in context")
}

// CtxActorUser is an alias for CtxTelegramUser (API keys populate the same context keys).
func CtxActorUser(c *gin.Context) (*store.TelegramUser, error) {
	return CtxTelegramUser(c)
}

// RateLimitKey returns a stable bucket key for per-actor rate limiting.
func RateLimitKey(c *gin.Context) string {
	if label, ok := c.Get(CtxKeyAPIKeyLabel); ok {
		if s, ok := label.(string); ok && s != "" {
			return "api:" + s
		}
	}
	if rawID, ok := c.Get(CtxKeyTelegramUserID); ok {
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
