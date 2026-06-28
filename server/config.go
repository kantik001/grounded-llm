package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config — настройки приложения из переменных окружения.
type Config struct {
	PythonRAGURL string // POST JSON → retrieval (контекст) в Python
	LLMAPIKey        string
	LLMModel         string
	LLMBaseURL       string
	ServerPort       string

	TelegramBotToken       string
	TelegramAuthDisabled   bool
	TelegramInitDataMaxAge time.Duration
	CORSAllowedOrigins     []string
	RateLimitPerMinute     int
	DatabaseURL            string
	UploadDir              string
	DataDir                string
	PythonBaseURL          string
	AdminUser              string
	AdminPassword          string
	AdminSecret            string
	DefaultTenantID        string
	DefaultLocale          string
}

var config *Config
var chatStore *ChatStore

// Загружает .env и собирает Config из переменных окружения.
func loadConfig() *Config {
	godotenv.Load()
	godotenv.Load("../.env")

	maxAgeSec, _ := strconv.Atoi(getEnv("TELEGRAM_INIT_DATA_MAX_AGE_SEC", "86400"))
	if maxAgeSec < 0 {
		maxAgeSec = 86400
	}
	rateLimit, _ := strconv.Atoi(getEnv("RATE_LIMIT_REQUESTS_PER_MINUTE", "30"))
	if rateLimit < 0 {
		rateLimit = 0
	}

	return &Config{
		PythonRAGURL: getEnv("PYTHON_RAG_URL", "http://python:5000/rag/context"),
		LLMAPIKey:        getEnv("LLM_API_KEY", ""),
		LLMBaseURL:       getEnv("LLM_BASE_URL", "https://openrouter.ai/api"),
		LLMModel:         getEnv("LLM_MODEL", "openrouter/free"),
		ServerPort:       getEnv("SERVER_PORT", "8080"),

		TelegramBotToken:       getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramAuthDisabled:   strings.EqualFold(getEnv("TELEGRAM_AUTH_DISABLED", "false"), "true"),
		TelegramInitDataMaxAge: time.Duration(maxAgeSec) * time.Second,
		CORSAllowedOrigins:     parseAllowedOrigins(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost,http://127.0.0.1")),
		RateLimitPerMinute:     rateLimit,
		DatabaseURL:            getEnv("DATABASE_URL", "postgres://grounded:grounded@postgres:5432/grounded?sslmode=disable"),
		UploadDir:              getEnv("UPLOAD_DIR", "/data/uploads"),
		DataDir:                getEnv("DATA_DIR", "/app/data"),
		PythonBaseURL:          getEnv("PYTHON_BASE_URL", "http://python:5000"),
		AdminUser:              getEnv("ADMIN_USER", "admin"),
		AdminPassword:          getEnv("ADMIN_PASSWORD", ""),
		AdminSecret:            getEnv("ADMIN_SECRET", ""),
		DefaultTenantID:        getEnv("DEFAULT_TENANT_ID", "default"),
		DefaultLocale:          getEnv("DEFAULT_LOCALE", "en"),
	}
}

// Возвращает значение переменной окружения или defaultValue.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Пишет в лог основные настройки при старте сервера.
func logStartup(cfg *Config) {
	log.Printf("Starting Grounded LLM Server...")
	log.Printf("Python RAG context URL: %s", cfg.PythonRAGURL)
	log.Printf("LLM Model: %s", cfg.LLMModel)
	if cfg.LLMAPIKey != "" {
		log.Printf("LLM API Key: configured")
	} else {
		log.Printf("LLM API Key: not configured")
	}
	if cfg.TelegramAuthDisabled {
		log.Printf("Telegram auth: DISABLED (dev mode only)")
	} else if cfg.TelegramBotToken != "" {
		log.Printf("Telegram auth: enabled")
	} else {
		log.Printf("Telegram auth: WARNING — TELEGRAM_BOT_TOKEN not set, protected routes will reject clients")
	}
	if len(apiKeyRegistry) > 0 {
		log.Printf("API keys: %d configured", len(apiKeyRegistry))
	}
	if len(adminUserRegistry) > 0 {
		log.Printf("Admin users (RBAC): %d from ADMIN_USERS_FILE", len(adminUserRegistry))
	} else if cfg.AdminPassword != "" {
		log.Printf("Admin auth: legacy single user %q (role: admin)", cfg.AdminUser)
	}
	if oidcConfigured() {
		log.Printf("OIDC SSO: enabled (issuer=%s)", oidcCfg.Issuer)
	}
	if len(tenantQuotaRegistry) > 0 {
		log.Printf("Tenant quotas: %d tenant(s)", len(tenantQuotaRegistry))
	}
	log.Printf("Default tenant: %s", cfg.DefaultTenantID)
	log.Printf("Default locale: %s", cfg.DefaultLocale)
	log.Printf("CORS origins: %v", cfg.CORSAllowedOrigins)
	log.Printf("Rate limit: %d req/min per user", cfg.RateLimitPerMinute)
	log.Printf("Database URL: configured")
	log.Printf("Upload dir: %s", cfg.UploadDir)
}
