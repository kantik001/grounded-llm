package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /branding — публичные тексты UI для Web App.
func handleBranding(c *gin.Context) {
	locale := ctxLocale(c)
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"locale":   locale,
		"branding": brandingForLocale(locale),
	})
}
