package config

import (
	"fmt"
	"os"
	"strings"
)

// IsProductionEnv reports whether the process should enforce production safety checks.
// Set GROUNDED_ENV=production (preferred) or APP_ENV=production.
func IsProductionEnv() bool {
	for _, key := range []string{"GROUNDED_ENV", "APP_ENV", "ENV"} {
		v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
		if v == "production" || v == "prod" {
			return true
		}
	}
	return false
}

func isProductionEnv() bool { return IsProductionEnv() }

// ValidateProduction fails fast on insecure production configuration.
// adminUserCount is the number of users loaded from ADMIN_USERS_FILE (0 if unused).
func ValidateProduction(cfg *Config, adminUserCount int) error {
	if cfg == nil || !IsProductionEnv() {
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
	if strings.TrimSpace(cfg.AdminPassword) == "" && adminUserCount == 0 {
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
