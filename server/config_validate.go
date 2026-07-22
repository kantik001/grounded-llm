package main

import (
	"fmt"
	"os"
	"strings"
)

// isProductionEnv reports whether the process should enforce production safety checks.
// Set GROUNDED_ENV=production (preferred) or APP_ENV=production.
func isProductionEnv() bool {
	for _, key := range []string{"GROUNDED_ENV", "APP_ENV", "ENV"} {
		v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
		if v == "production" || v == "prod" {
			return true
		}
	}
	return false
}

// validateProductionConfig fails fast on insecure production configuration.
func validateProductionConfig(cfg *Config) error {
	if cfg == nil || !isProductionEnv() {
		return nil
	}
	var problems []string

	if cfg.TelegramAuthDisabled {
		problems = append(problems, "TELEGRAM_AUTH_DISABLED=true is not allowed in production")
	}
	if cfg.LLMMock {
		problems = append(problems, "LLM_MOCK=true is not allowed in production")
	}
	if cfg.RAGMock {
		problems = append(problems, "RAG_MOCK=true is not allowed in production")
	}
	if strings.TrimSpace(cfg.AdminPassword) == "" && len(adminUserRegistry) == 0 {
		problems = append(problems, "ADMIN_PASSWORD (or ADMIN_USERS_FILE) must be set in production")
	}
	if strings.TrimSpace(cfg.RAGServiceToken) == "" {
		problems = append(problems, "RAG_SERVICE_TOKEN must be set in production (Go ↔ Python internal auth)")
	}
	if strings.TrimSpace(cfg.AdminSecret) == "" {
		problems = append(problems, "ADMIN_SECRET must be set in production (Python admin routes)")
	}
	if strings.Contains(cfg.DatabaseURL, "grounded:grounded@") {
		problems = append(problems, "default Postgres password grounded:grounded is not allowed in production")
	}
	if len(cfg.CORSAllowedOrigins) == 1 && cfg.CORSAllowedOrigins[0] == "*" {
		problems = append(problems, "CORS_ALLOWED_ORIGINS=* is not allowed in production")
	}
	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("production safety check failed:\n  - %s", strings.Join(problems, "\n  - "))
}
