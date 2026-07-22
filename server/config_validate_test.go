package main

import "testing"

func TestIsProductionEnv(t *testing.T) {
	t.Setenv("GROUNDED_ENV", "")
	t.Setenv("APP_ENV", "")
	t.Setenv("ENV", "")
	if isProductionEnv() {
		t.Fatal("expected false when unset")
	}
	t.Setenv("GROUNDED_ENV", "production")
	if !isProductionEnv() {
		t.Fatal("expected true for GROUNDED_ENV=production")
	}
}

func TestValidateProductionConfig_DevSkips(t *testing.T) {
	t.Setenv("GROUNDED_ENV", "development")
	cfg := &Config{TelegramAuthDisabled: true}
	if err := validateProductionConfig(cfg); err != nil {
		t.Fatalf("dev should skip: %v", err)
	}
}

func TestValidateProductionConfig_FailsOnInsecure(t *testing.T) {
	t.Setenv("GROUNDED_ENV", "production")
	adminUserRegistry = nil
	cfg := &Config{
		TelegramAuthDisabled: true,
		LLMMock:              true,
		RAGMock:              true,
		AdminPassword:        "",
		RAGServiceToken:      "",
		AdminSecret:          "",
		DatabaseURL:          "postgres://grounded:grounded@postgres:5432/grounded?sslmode=disable",
		CORSAllowedOrigins:   []string{"*"},
	}
	if err := validateProductionConfig(cfg); err == nil {
		t.Fatal("expected production validation error")
	}
}

func TestValidateProductionConfig_OK(t *testing.T) {
	t.Setenv("GROUNDED_ENV", "production")
	adminUserRegistry = nil
	cfg := &Config{
		TelegramAuthDisabled: false,
		LLMMock:              false,
		RAGMock:              false,
		AdminPassword:        "strong-password-here",
		RAGServiceToken:      "token-32-chars-minimum-xxxxxxxxxx",
		AdminSecret:          "admin-secret-value",
		DatabaseURL:          "postgres://app:s3cret@db:5432/grounded?sslmode=require",
		CORSAllowedOrigins:   []string{"https://app.example.com"},
	}
	if err := validateProductionConfig(cfg); err != nil {
		t.Fatalf("expected OK: %v", err)
	}
}
