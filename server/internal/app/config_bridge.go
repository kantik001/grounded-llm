package app

import (
	"fmt"
	"log"

	"grounded_llm_server/internal/admin"
	cfgpkg "grounded_llm_server/internal/config"
	"grounded_llm_server/internal/oidc"
	"grounded_llm_server/internal/tenant"
)

// Config is the process configuration (owned by internal/config).
type Config = cfgpkg.Config

func loadConfig() *Config {
	return cfgpkg.Load()
}

func validateProductionConfig(cfg *Config) error {
	if err := cfgpkg.ValidateProduction(cfg, admin.UserCount()); err != nil {
		return err
	}
	if cfgpkg.IsProductionEnv() {
		if n := admin.PlaintextPasswordCount(); n > 0 {
			return fmt.Errorf("production safety check failed: %d admin user(s) in ADMIN_USERS_FILE use plaintext \"password\" — switch to \"password_bcrypt\"", n)
		}
	}
	return nil
}

func logStartup(cfg *Config) {
	cfgpkg.LogBasic(cfg)
	if len(apiKeyRegistry) > 0 {
		log.Printf("API keys: %d configured", len(apiKeyRegistry))
	}
	if admin.UserCount() > 0 {
		log.Printf("Admin users (RBAC): %d from ADMIN_USERS_FILE", admin.UserCount())
	} else if cfg.AdminPassword != "" {
		log.Printf("Admin auth: legacy single user %q (role: admin)", cfg.AdminUser)
	}
	if oidcConfigured() {
		log.Printf("OIDC SSO: enabled (issuer=%s)", oidc.Issuer())
	}
	if n := tenant.ConfiguredQuotaCount(); n > 0 {
		log.Printf("Tenant quotas: %d tenant(s)", n)
	}
	if saasSignupEnabled() {
		log.Printf("SaaS signup: enabled")
	}
	if stripeWebhookSecret() != "" {
		log.Printf("Stripe webhook: configured")
	}
	if stripeSecretKey() != "" {
		log.Printf("Stripe checkout: configured")
	}
}
