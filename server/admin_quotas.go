package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GET /admin/quotas?tenant_id=
func handleAdminQuotas(c *gin.Context) {
	tenantID := adminTenantID(c)
	if raw := strings.TrimSpace(c.Query("tenant_id")); raw != "" {
		tenantID = normalizeTenantID(raw)
	}
	status, err := buildTenantQuotaStatus(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"quota":   status,
	})
}

func quotaErrorResponse(c *gin.Context, err error) {
	c.JSON(http.StatusTooManyRequests, gin.H{
		"success": false,
		"error":   err.Error(),
		"code":    "quota_exceeded",
	})
}
