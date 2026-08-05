package locale

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	cfgpkg "grounded_llm_server/internal/config"
)

const (
	CtxKeyLocale   = "locale"
	HeaderLocale   = "X-Locale"
	fallbackLocale = "en"
)

var appDefaultLocale = "en"

var supportedLocales = []string{"ru", "en"}

type localeBundle struct {
	prompts    map[string]DomainPrompts
	platform   platformPrompts
	onboarding map[string][]string
	branding   BrandingConfig
}

var localeBundles map[string]*localeBundle

// DomainPrompts holds RAG prompt strings for one knowledge domain.
type DomainPrompts struct {
	RAGSystem    string `json:"rag_system"`
	RAGTaskIntro string `json:"rag_task_intro"`
}

type platformPrompts struct {
	RAGConstraints string `json:"rag_constraints"`
	VerifyFailHint string `json:"verify_fail_hint"`
}

// BrandingConfig holds public UI copy for a locale bundle.
type BrandingConfig struct {
	AppTitle           string `json:"app_title"`
	HeaderEmoji        string `json:"header_emoji"`
	HeaderSubtitle     string `json:"header_subtitle"`
	DomainLabel        string `json:"domain_label"`
	OnboardingTitle    string `json:"onboarding_title"`
	ChatDivider        string `json:"chat_divider"`
	Disclaimer         string `json:"disclaimer"`
	PageTitleSuffix    string `json:"page_title_suffix"`
	TypingLine         string `json:"typing_line"`
	MessagePlaceholder string `json:"message_placeholder"`
	SendAriaLabel      string `json:"send_aria_label"`
	DomainSelectAria   string `json:"domain_select_aria"`
	EmptyChatHint      string `json:"empty_chat_hint"`
	CitationsTitle     string `json:"citations_title"`
	DocumentFallback   string `json:"document_fallback"`
	PagePrefix         string `json:"page_prefix"`
	DomainComingSoon   string `json:"domain_coming_soon"`
	UserImageAlt       string `json:"user_image_alt"`
}

// Init loads locale bundles and sets the process default locale.
func Init(cfg *cfgpkg.Config) {
	appDefaultLocale = "en"
	if cfg != nil && cfg.DefaultLocale != "" {
		appDefaultLocale = BundleLocale(cfg.DefaultLocale)
	}
	if l := NormalizeLocale(os.Getenv("DEFAULT_LOCALE")); l != "" {
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

// DefaultLocale returns the configured process default locale code.
func DefaultLocale() string {
	return appDefaultLocale
}

func NormalizeLocale(raw string) string {
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
		filepath.Join("config", "locales", locale, name),
		filepath.Join("..", "config", "locales", locale, name),
		filepath.Join("..", "..", "config", "locales", locale, name),
		filepath.Join("..", "..", "..", "config", "locales", locale, name),
	} {
		if _, err := os.Stat(base); err == nil {
			return base
		}
	}
	return filepath.Join("config", "locales", locale, name)
}

func loadLocaleBundle(locale string) error {
	b := &localeBundle{
		prompts:    make(map[string]DomainPrompts),
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
		var p DomainPrompts
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

// ReloadBundles re-reads locale JSON from disk (SIGHUP / periodic reload).
func ReloadBundles() error {
	return loadAllLocaleBundles()
}

// ResolveLocale picks the request locale and stores it on the Gin context.
func ResolveLocale(c *gin.Context) string {
	if v, ok := c.Get(CtxKeyLocale); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	if h := NormalizeLocale(c.GetHeader(HeaderLocale)); h != "" {
		c.Set(CtxKeyLocale, h)
		return h
	}
	if q := NormalizeLocale(c.Query("locale")); q != "" {
		c.Set(CtxKeyLocale, q)
		return q
	}
	if al := c.GetHeader("Accept-Language"); al != "" {
		for _, part := range strings.Split(al, ",") {
			tag := strings.TrimSpace(strings.Split(part, ";")[0])
			if loc := NormalizeLocale(tag); loc != "" {
				c.Set(CtxKeyLocale, loc)
				return loc
			}
		}
	}
	if lang := telegramLanguageFromContext(c); lang != "" {
		if loc := NormalizeLocale(lang); loc != "" {
			c.Set(CtxKeyLocale, loc)
			return loc
		}
	}
	c.Set(CtxKeyLocale, appDefaultLocale)
	return appDefaultLocale
}

// CtxLocale returns the resolved locale for a request, falling back to the default bundle.
func CtxLocale(c *gin.Context) string {
	if v, ok := c.Get(CtxKeyLocale); ok {
		if s, ok := v.(string); ok && s != "" {
			return BundleLocale(s)
		}
	}
	return BundleLocale(appDefaultLocale)
}

// BundleLocale maps a locale code to a loaded bundle key.
func BundleLocale(locale string) string {
	if _, ok := localeBundles[locale]; ok {
		return locale
	}
	if _, ok := localeBundles[fallbackLocale]; ok {
		return fallbackLocale
	}
	return appDefaultLocale
}

// Middleware resolves locale on each request.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ResolveLocale(c)
		c.Next()
	}
}

// PromptsForDomainLocale returns RAG prompts for a domain and locale.
func PromptsForDomainLocale(domainID, locale string) DomainPrompts {
	b := localeBundles[BundleLocale(locale)]
	if b == nil {
		return DomainPrompts{}
	}
	if p, ok := b.prompts[domainID]; ok {
		return p
	}
	if p, ok := b.prompts[defaultDomainID()]; ok {
		return p
	}
	return DomainPrompts{
		RAGSystem:    "You are a document assistant. Answer only from context.",
		RAGTaskIntro: "Answer strictly from the provided context.",
	}
}

// RAGConstraintsForLocale returns platform RAG constraint text.
func RAGConstraintsForLocale(locale string) string {
	b := localeBundles[BundleLocale(locale)]
	if b != nil && b.platform.RAGConstraints != "" {
		return b.platform.RAGConstraints
	}
	return `- Do not invent facts. If the answer is not in the context, say so clearly.`
}

// VerifyFailHintForLocale returns the hint shown when RAG verification fails.
func VerifyFailHintForLocale(locale string) string {
	b := localeBundles[BundleLocale(locale)]
	if b != nil && b.platform.VerifyFailHint != "" {
		return b.platform.VerifyFailHint
	}
	return "Contact your knowledge base administrator."
}

// BrandingForLocale returns public UI branding strings.
func BrandingForLocale(locale string) BrandingConfig {
	b := localeBundles[BundleLocale(locale)]
	if b != nil {
		return b.branding
	}
	return BrandingConfig{}
}

// OnboardingForDomainLocale returns onboarding question strings.
func OnboardingForDomainLocale(domainID, locale string) []string {
	b := localeBundles[BundleLocale(locale)]
	if b == nil {
		return nil
	}
	if q, ok := b.onboarding[domainID]; ok {
		return q
	}
	return b.onboarding[defaultDomainID()]
}
