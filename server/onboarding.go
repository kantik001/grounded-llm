package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleOnboarding(c *gin.Context) {
	locale := ctxLocale(c)
	domainID, err := normalizeDomainID(domainIDFromQuery(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	questions := onboardingForDomainLocale(domainID, locale)
	if questions == nil {
		questions = []string{}
	}
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"locale":     locale,
		"domain_id":  domainID,
		"questions":  questions,
	})
}
