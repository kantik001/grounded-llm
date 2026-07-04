package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

const auditActionTenantPurge = "tenant_purge"

var validTenantID = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,62}$`)

// TenantPurgeStats counts removed records and files.
type TenantPurgeStats struct {
	Sessions      int64 `json:"sessions"`
	Messages      int64 `json:"messages"`
	FeedbackRows  int64 `json:"feedback_rows"`
	AuditRows     int64 `json:"audit_rows"`
	AnalyticsRows int64 `json:"analytics_rows"`
	ReindexJobs   int64 `json:"reindex_jobs"`
	DataFiles     int   `json:"data_files"`
	UploadTokens  int64 `json:"upload_tokens"`
}

func validateTenantPurgeTarget(tenantID string, confirm, purgeDefault bool) error {
	tenantID = normalizeTenantID(tenantID)
	if tenantID == "" {
		return fmt.Errorf("tenant_id required")
	}
	if !validTenantID.MatchString(tenantID) {
		return fmt.Errorf("invalid tenant_id")
	}
	if !confirm {
		return fmt.Errorf("confirm=true required")
	}
	if tenantID == normalizeTenantID(config.DefaultTenantID) && !purgeDefault {
		return fmt.Errorf("purging default tenant requires purge_default=true")
	}
	return nil
}

// DELETE /admin/tenants/:tenant_id?confirm=true&purge_default=false
func handleAdminPurgeTenant(c *gin.Context) {
	if chatStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Store not ready"})
		return
	}
	tenantID := normalizeTenantID(c.Param("tenant_id"))
	confirm := strings.EqualFold(strings.TrimSpace(c.Query("confirm")), "true")
	purgeDefault := strings.EqualFold(strings.TrimSpace(c.Query("purge_default")), "true")
	if err := validateTenantPurgeTarget(tenantID, confirm, purgeDefault); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	ctx := c.Request.Context()
	active, err := chatStore.HasActiveReindexJob(ctx, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if active {
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": "reindex job in progress for tenant"})
		return
	}
	stats, err := chatStore.PurgeTenant(ctx, config.DataDir, tenantID)
	if err != nil {
		recordAdminAudit(c, auditOpts{
			Action:   auditActionTenantPurge,
			TenantID: tenantID,
			Success:  false,
			Details:  map[string]any{"error": err.Error()},
		})
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	recordAdminAudit(c, auditOpts{
		Action:   auditActionTenantPurge,
		TenantID: tenantID,
		Success:  true,
		Details: map[string]any{
			"sessions":  stats.Sessions,
			"messages":  stats.Messages,
			"data_files": stats.DataFiles,
		},
	})
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"tenant_id": tenantID,
		"deleted":   stats,
	})
}

func countDataFiles(dir string) int {
	n := 0
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if isKnowledgeFile(d.Name()) {
			n++
		}
		return nil
	})
	return n
}
