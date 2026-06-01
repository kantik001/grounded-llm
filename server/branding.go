package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// BrandingConfig — тексты Web App (domain pack, не core).
type BrandingConfig struct {
	AppTitle         string `json:"app_title"`
	HeaderEmoji      string `json:"header_emoji"`
	HeaderSubtitle   string `json:"header_subtitle"`
	DomainLabel      string `json:"domain_label"`
	CropLabel        string `json:"crop_label"` // legacy
	OnboardingTitle  string `json:"onboarding_title"`
	ChatDivider      string `json:"chat_divider"`
	Disclaimer       string `json:"disclaimer"`
}

var brandingCatalog BrandingConfig

func loadBrandingConfig() error {
	path := brandingConfigPath()
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, &brandingCatalog)
}

func brandingConfigPath() string {
	if p := os.Getenv("BRANDING_CONFIG_PATH"); p != "" {
		return p
	}
	for _, candidate := range []string{
		"/config/branding.json",
		filepath.Join("..", "config", "branding.json"),
		filepath.Join("config", "branding.json"),
	} {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return filepath.Join("config", "branding.json")
}

// GET /branding — публичные тексты UI для Web App.
func handleBranding(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"branding": brandingCatalog,
	})
}
