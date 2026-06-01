package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

var onboardingQuestions map[string][]string

func loadOnboardingConfig() error {
	path := onboardingConfigPath()
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, &onboardingQuestions)
}

func onboardingConfigPath() string {
	if p := os.Getenv("ONBOARDING_CONFIG_PATH"); p != "" {
		return p
	}
	for _, candidate := range []string{
		"/config/onboarding.json",
		filepath.Join("..", "config", "onboarding.json"),
		filepath.Join("config", "onboarding.json"),
	} {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return filepath.Join("config", "onboarding.json")
}

func handleOnboarding(c *gin.Context) {
	raw := strings.TrimSpace(c.Query("domain_id"))
	if raw == "" {
		raw = strings.TrimSpace(c.Query("crop_id"))
	}
	domainID, err := normalizeDomainID(raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	questions := onboardingQuestions[domainID]
	if questions == nil {
		questions = []string{}
	}
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"domain_id":  domainID,
		"questions":  questions,
	})
}
