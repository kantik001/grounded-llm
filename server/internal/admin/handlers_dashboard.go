package admin

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/audit"
	"grounded_llm_server/internal/store"
	"grounded_llm_server/internal/tenant"
)

func handleAnalytics(c *gin.Context) {
	s := st()
	if s == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Database unavailable"})
		return
	}
	tenantID := strings.TrimSpace(c.Query("tenant_id"))
	days := store.ParseAnalyticsDays(c.Query("days"))
	dash, err := s.AnalyticsDashboard(c.Request.Context(), tenantID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "analytics": dash})
}

func handleFeedbackSummary(c *gin.Context) {
	s := st()
	if s == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Database unavailable"})
		return
	}
	summary, err := s.FeedbackSummary(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "feedback": summary})
}

func handleAuditLog(c *gin.Context) {
	s := st()
	if s == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Store not ready"})
		return
	}
	limit, offset, action := audit.ParseLogQuery(c)
	entries, err := s.ListAuditLog(c.Request.Context(), limit, offset, action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if entries == nil {
		entries = []store.AuditEntry{}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"entries": entries,
		"limit":   limit,
		"offset":  offset,
	})
}

func handleQuotas(c *gin.Context) {
	tenantID := tenant.AdminID(c)
	if raw := strings.TrimSpace(c.Query("tenant_id")); raw != "" {
		tenantID = tenant.NormalizeTenantID(raw)
	}
	status, err := tenant.BuildQuotaStatus(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"quota":   status,
	})
}
