package config

import "testing"

func TestIsProductionEnv(t *testing.T) {
	t.Setenv("GROUNDED_ENV", "")
	t.Setenv("APP_ENV", "")
	t.Setenv("ENV", "")
	if IsProductionEnv() {
		t.Fatal("expected false when unset")
	}
	t.Setenv("GROUNDED_ENV", "production")
	if !IsProductionEnv() {
		t.Fatal("expected true for GROUNDED_ENV=production")
	}
}

func TestValidateProduction_DevSkips(t *testing.T) {
	t.Setenv("GROUNDED_ENV", "development")
	cfg := &Config{TelegramAuthDisabled: true}
	if err := ValidateProduction(cfg, 0); err != nil {
		t.Fatalf("dev should skip: %v", err)
	}
}

func TestValidateProduction_FailsOnInsecure(t *testing.T) {
	t.Setenv("GROUNDED_ENV", "production")
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
	if err := ValidateProduction(cfg, 0); err == nil {
		t.Fatal("expected production validation error")
	}
}

func TestValidateProduction_OK(t *testing.T) {
	t.Setenv("GROUNDED_ENV", "production")
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
	if err := ValidateProduction(cfg, 0); err != nil {
		t.Fatalf("expected OK: %v", err)
	}
}
