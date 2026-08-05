package app

import (
	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/auth"
	cfgpkg "grounded_llm_server/internal/config"
	"grounded_llm_server/internal/locale"
)

func init() {
	locale.BindConfig(func() *cfgpkg.Config { return config })
	locale.BindTelegramLanguageFromContext(func(c *gin.Context) string {
		if u, err := auth.CtxTelegramUser(c); err == nil && u != nil {
			return u.LanguageCode
		}
		return ""
	})
}

type (
	domainPrompts  = locale.DomainPrompts
	BrandingConfig = locale.BrandingConfig
)

const (
	ctxKeyLocale = locale.CtxKeyLocale
	headerLocale = locale.HeaderLocale
)

func initLocaleConfig(cfg *Config) {
	locale.Init(cfg)
}

func normalizeLocale(raw string) string {
	return locale.NormalizeLocale(raw)
}

func resolveLocale(c *gin.Context, cfg *Config) string {
	_ = cfg
	return locale.ResolveLocale(c)
}

func ctxLocale(c *gin.Context) string {
	return locale.CtxLocale(c)
}

func bundleLocale(localeCode string) string {
	return locale.BundleLocale(localeCode)
}

func localeMiddleware(cfg *Config) gin.HandlerFunc {
	_ = cfg
	return locale.Middleware()
}

func promptsForDomainLocale(domainID, localeCode string) domainPrompts {
	return locale.PromptsForDomainLocale(domainID, localeCode)
}

func ragConstraintsForLocale(localeCode string) string {
	return locale.RAGConstraintsForLocale(localeCode)
}

func verifyFailHintForLocale(localeCode string) string {
	return locale.VerifyFailHintForLocale(localeCode)
}

func brandingForLocale(localeCode string) BrandingConfig {
	return locale.BrandingForLocale(localeCode)
}

func onboardingForDomainLocale(domainID, localeCode string) []string {
	return locale.OnboardingForDomainLocale(domainID, localeCode)
}

func reloadLocaleBundles() error {
	return locale.ReloadBundles()
}
