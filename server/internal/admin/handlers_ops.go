package admin

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/audit"
	"grounded_llm_server/internal/auth"
	"grounded_llm_server/internal/store"
	"grounded_llm_server/internal/tenant"
)

const auditActionTenantPurge = "tenant_purge"

var validTenantID = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,62}$`)

type apiKeySummary struct {
	Label string   `json:"label"`
	Roles []string `json:"roles"`
}

func handleAPIKeys(c *gin.Context) {
	keys := listAPIKeySummaries()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"keys":    keys,
		"users":   listUserSummaries(),
		"roles":   AllRoles,
	})
}

func listAPIKeySummaries() []apiKeySummary {
	out := make([]apiKeySummary, 0, len(auth.Registry))
	for _, rec := range auth.Registry {
		roles := rec.Roles
		if len(roles) == 0 {
			roles = DefaultAPIKeyRoles()
		}
		out = append(out, apiKeySummary{Label: rec.Label, Roles: roles})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Label < out[j].Label })
	return out
}

func handleReindex(c *gin.Context) {
	s := st()
	if s == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Database unavailable"})
		return
	}
	actor := audit.ActorFromContext(c)
	tenantID := tenant.AdminID(c)
	domainID := c.Query("domain_id")
	if domainID == "" {
		domainID = c.PostForm("domain_id")
	}

	job, alreadyRunning, err := s.CreateReindexJob(c.Request.Context(), actor, tenantID, domainID)
	if err != nil {
		log.Printf("CreateReindexJob: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !alreadyRunning {
		startReindexWorker(job)
	}
	c.JSON(http.StatusAccepted, gin.H{
		"success":         true,
		"job_id":          job.ID,
		"status":          job.Status,
		"status_label":    reindexStatusLabel(job.Status),
		"already_running": alreadyRunning,
		"message":         reindexAcceptedMessage(alreadyRunning),
	})
}

func handleReindexStatus(c *gin.Context) {
	s := st()
	if s == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Database unavailable"})
		return
	}
	ctx := c.Request.Context()
	var job *store.ReindexJob
	var err error

	if raw := c.Query("job_id"); raw != "" {
		id, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid job_id"})
			return
		}
		job, err = s.GetReindexJob(ctx, id)
	} else {
		job, err = s.ActiveReindexJob(ctx)
		if err == nil && job == nil {
			job, err = s.GetLatestReindexJob(ctx)
		}
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "No reindex job found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"job":          job,
		"status_label": reindexStatusLabel(job.Status),
		"done":         isReindexTerminal(job.Status),
	})
}

func validateTenantPurgeTarget(tenantID string, confirm, purgeDefault bool) error {
	tenantID = tenant.NormalizeTenantID(tenantID)
	if tenantID == "" {
		return fmt.Errorf("tenant_id required")
	}
	if !validTenantID.MatchString(tenantID) {
		return fmt.Errorf("invalid tenant_id")
	}
	if !confirm {
		return fmt.Errorf("confirm=true required")
	}
	c := cfg()
	defaultTenant := "default"
	if c != nil {
		defaultTenant = tenant.NormalizeTenantID(c.DefaultTenantID)
	}
	if tenantID == defaultTenant && !purgeDefault {
		return fmt.Errorf("purging default tenant requires purge_default=true")
	}
	return nil
}

func handlePurgeTenant(c *gin.Context) {
	s := st()
	if s == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Store not ready"})
		return
	}
	tenantID := tenant.NormalizeTenantID(c.Param("tenant_id"))
	confirm := strings.EqualFold(strings.TrimSpace(c.Query("confirm")), "true")
	purgeDefault := strings.EqualFold(strings.TrimSpace(c.Query("purge_default")), "true")
	if err := validateTenantPurgeTarget(tenantID, confirm, purgeDefault); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	ctx := c.Request.Context()
	active, err := s.HasActiveReindexJob(ctx, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if active {
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": "reindex job in progress for tenant"})
		return
	}
	cCfg := cfg()
	dataDir := ""
	if cCfg != nil {
		dataDir = cCfg.DataDir
	}
	stats, err := s.PurgeTenant(ctx, dataDir, tenantID)
	if err != nil {
		audit.Record(c, audit.Opts{
			Action:   auditActionTenantPurge,
			TenantID: tenantID,
			Success:  false,
			Details:  map[string]any{"error": err.Error()},
		})
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	audit.Record(c, audit.Opts{
		Action:   auditActionTenantPurge,
		TenantID: tenantID,
		Success:  true,
		Details: map[string]any{
			"sessions":   stats.Sessions,
			"messages":   stats.Messages,
			"data_files": stats.DataFiles,
		},
	})
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"tenant_id": tenantID,
		"deleted":   stats,
	})
}

// ValidateTenantPurgeTarget is exported for tests.
func ValidateTenantPurgeTarget(tenantID string, confirm, purgeDefault bool) error {
	return validateTenantPurgeTarget(tenantID, confirm, purgeDefault)
}
