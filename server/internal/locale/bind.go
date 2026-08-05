package locale

import (
	"github.com/gin-gonic/gin"

	cfgpkg "grounded_llm_server/internal/config"
)

var (
	lookupDefaultDomainID         func() string
	lookupTelegramLanguageFromCtx func(c *gin.Context) string
)

// BindConfig is kept for app wiring compatibility; locale reads config via Init().
func BindConfig(fn func() *cfgpkg.Config) {
	_ = fn
}

// BindDefaultDomainID wires the default knowledge domain ID (from internal/domain).
func BindDefaultDomainID(fn func() string) {
	lookupDefaultDomainID = fn
}

func defaultDomainID() string {
	if lookupDefaultDomainID != nil {
		return lookupDefaultDomainID()
	}
	return "default"
}

// BindTelegramLanguageFromContext resolves Accept-Language fallback from Telegram user context.
func BindTelegramLanguageFromContext(fn func(c *gin.Context) string) {
	lookupTelegramLanguageFromCtx = fn
}

func telegramLanguageFromContext(c *gin.Context) string {
	if lookupTelegramLanguageFromCtx != nil {
		return lookupTelegramLanguageFromCtx(c)
	}
	return ""
}
