package locale

import (
	"github.com/gin-gonic/gin"

	cfgpkg "grounded_llm_server/internal/config"
)

var lookupConfig = cfgpkg.Current

var (
	lookupDefaultDomainID         func() string
	lookupTelegramLanguageFromCtx func(c *gin.Context) string
)

// BindConfig overrides where this package reads process config (app sets app.config).
func BindConfig(fn func() *cfgpkg.Config) {
	if fn != nil {
		lookupConfig = fn
	}
}

func cfg() *cfgpkg.Config {
	return lookupConfig()
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
