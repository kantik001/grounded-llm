package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ctxKeyLocale   = "locale"
	headerLocale   = "X-Locale"
	fallbackLocale = "en"
)

var appDefaultLocale = "ru"

var supportedLocales = []string{"ru", "en"}

type localeBundle struct {
	prompts    map[string]domainPrompts
	platform   platformPrompts
	onboarding map[string][]string
	branding   BrandingConfig
}

var localeBundles map[string]*localeBundle

type domainPrompts struct {
	RAGSystem    string `json:"rag_system"`
	RAGTaskIntro string `json:"rag_task_intro"`
}

type platformPrompts struct {
	RAGConstraints string `json:"rag_constraints"`
	VerifyFailHint string `json:"verify_fail_hint"`
}

type BrandingConfig struct {
	AppTitle        string `json:"app_title"`
	HeaderEmoji     string `json:"header_emoji"`
	HeaderSubtitle  string `json:"header_subtitle"`
	DomainLabel     string `json:"domain_label"`
	OnboardingTitle string `json:"onboarding_title"`
	ChatDivider     string `json:"chat_divider"`
	Disclaimer      string `json:"disclaimer"`
}

func initLocaleConfig(cfg *Config) {
	appDefaultLocale = "ru"
	if cfg != nil && cfg.DefaultLocale != "" {
		appDefaultLocale = bundleLocale(cfg.DefaultLocale)
	}
	if l := normalizeLocale(os.Getenv("DEFAULT_LOCALE")); l != "" {
		appDefaultLocale = l
	}
	localeBundles = make(map[string]*localeBundle)
	if err := loadAllLocaleBundles(); err != nil {
		panic(err)
	}
	if _, ok := localeBundles[appDefaultLocale]; !ok {
		panic("default locale bundle not loaded: " + appDefaultLocale)
	}
}

func normalizeLocale(raw string) string {
	l := strings.ToLower(strings.TrimSpace(raw))
	if l == "ru" || strings.HasPrefix(l, "ru-") {
		return "ru"
	}
	if l == "en" || strings.HasPrefix(l, "en-") {
		return "en"
	}
	return ""
}

func localeConfigPath(locale, name string) string {
	if p := strings.TrimSpace(os.Getenv("LOCALES_ROOT")); p != "" {
		return filepath.Join(p, locale, name)
	}
	for _, base := range []string{
		filepath.Join("/config", "locales", locale, name),
		filepath.Join("..", "config", "locales", locale, name),
		filepath.Join("config", "locales", locale, name),
	} {
		if _, err := os.Stat(base); err == nil {
			return base
		}
	}
	return filepath.Join("config", "locales", locale, name)
}

func loadLocaleBundle(locale string) error {
	b := &localeBundle{
		prompts:    make(map[string]domainPrompts),
		onboarding: make(map[string][]string),
	}
	body, err := os.ReadFile(localeConfigPath(locale, "prompts.json"))
	if err != nil {
		return fmt.Errorf("prompts: %w", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return err
	}
	for key, val := range raw {
		if key == "_platform" {
			if err := json.Unmarshal(val, &b.platform); err != nil {
				return err
			}
			continue
		}
		var p domainPrompts
		if err := json.Unmarshal(val, &p); err != nil {
			return err
		}
		b.prompts[key] = p
	}
	if err := json.Unmarshal(readFileOrEmpty(locale, "onboarding.json"), &b.onboarding); err != nil {
		return fmt.Errorf("onboarding: %w", err)
	}
	if err := json.Unmarshal(readFileOrEmpty(locale, "branding.json"), &b.branding); err != nil {
		return fmt.Errorf("branding: %w", err)
	}
	localeBundles[locale] = b
	return nil
}

func readFileOrEmpty(locale, name string) []byte {
	body, err := os.ReadFile(localeConfigPath(locale, name))
	if err != nil {
		return []byte("{}")
	}
	return body
}

func loadAllLocaleBundles() error {
	for _, loc := range supportedLocales {
		if err := loadLocaleBundle(loc); err != nil {
			return fmt.Errorf("%s: %w", loc, err)
		}
	}
	return nil
}

func resolveLocale(c *gin.Context, cfg *Config) string {
	if v, ok := c.Get(ctxKeyLocale); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	if h := normalizeLocale(c.GetHeader(headerLocale)); h != "" {
		c.Set(ctxKeyLocale, h)
		return h
	}
	if q := normalizeLocale(c.Query("locale")); q != "" {
		c.Set(ctxKeyLocale, q)
		return q
	}
	if al := c.GetHeader("Accept-Language"); al != "" {
		for _, part := range strings.Split(al, ",") {
			tag := strings.TrimSpace(strings.Split(part, ";")[0])
			if loc := normalizeLocale(tag); loc != "" {
				c.Set(ctxKeyLocale, loc)
				return loc
			}
		}
	}
	if u, err := ctxTelegramUser(c); err == nil && u != nil {
		if loc := normalizeLocale(u.LanguageCode); loc != "" {
			c.Set(ctxKeyLocale, loc)
			return loc
		}
	}
	c.Set(ctxKeyLocale, appDefaultLocale)
	return appDefaultLocale
}

func ctxLocale(c *gin.Context) string {
	if v, ok := c.Get(ctxKeyLocale); ok {
		if s, ok := v.(string); ok && s != "" {
			return bundleLocale(s)
		}
	}
	return bundleLocale(appDefaultLocale)
}

func bundleLocale(locale string) string {
	if _, ok := localeBundles[locale]; ok {
		return locale
	}
	if _, ok := localeBundles[fallbackLocale]; ok {
		return fallbackLocale
	}
	return appDefaultLocale
}

func localeMiddleware(cfg *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		resolveLocale(c, cfg)
		c.Next()
	}
}

func promptsForDomainLocale(domainID, locale string) domainPrompts {
	b := localeBundles[bundleLocale(locale)]
	if b == nil {
		return domainPrompts{}
	}
	if p, ok := b.prompts[domainID]; ok {
		return p
	}
	if p, ok := b.prompts[defaultDomainID()]; ok {
		return p
	}
	return domainPrompts{
		RAGSystem:    "You are a document assistant. Answer only from context.",
		RAGTaskIntro: "Answer strictly from the provided context.",
	}
}

func ragConstraintsForLocale(locale string) string {
	b := localeBundles[bundleLocale(locale)]
	if b != nil && b.platform.RAGConstraints != "" {
		return b.platform.RAGConstraints
	}
	return `- Do not invent facts. If the answer is not in the context, say so clearly.`
}

func verifyFailHintForLocale(locale string) string {
	b := localeBundles[bundleLocale(locale)]
	if b != nil && b.platform.VerifyFailHint != "" {
		return b.platform.VerifyFailHint
	}
	return "Contact your knowledge base administrator."
}

func brandingForLocale(locale string) BrandingConfig {
	b := localeBundles[bundleLocale(locale)]
	if b != nil {
		return b.branding
	}
	return BrandingConfig{}
}

func onboardingForDomainLocale(domainID, locale string) []string {
	b := localeBundles[bundleLocale(locale)]
	if b == nil {
		return nil
	}
	if q, ok := b.onboarding[domainID]; ok {
		return q
	}
	return b.onboarding[defaultDomainID()]
}
