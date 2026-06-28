package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GET /admin/analytics — questions/day, verify pass rate, KB gaps, feedback.
func handleAdminAnalytics(c *gin.Context) {
	if chatStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Database unavailable"})
		return
	}
	tenantID := strings.TrimSpace(c.Query("tenant_id"))
	days := parseAnalyticsDays(c.Query("days"))
	dash, err := chatStore.AnalyticsDashboard(c.Request.Context(), tenantID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "analytics": dash})
}
