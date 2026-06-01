package main

import (
	"encoding/json"
	"net/http"
	"os"

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
	return resolveConfigPath("ONBOARDING_CONFIG_PATH", defaultConfigCandidates("onboarding.json")...)
}

func handleOnboarding(c *gin.Context) {
	domainID, err := normalizeDomainID(domainIDFromQuery(c))
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
