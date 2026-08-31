package admin

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/audit"
	"grounded_llm_server/internal/domain"
	"grounded_llm_server/internal/store"
	"grounded_llm_server/internal/tenant"
)

type ingestRequest struct {
	Files               []string `json:"files"`
	Mode                string   `json:"mode"`
	Source              string   `json:"source"`
	Sync                bool     `json:"sync"`
	IndexRunID          string   `json:"index_run_id"`
	ActivateOnComplete  bool     `json:"activate_on_complete"`
}

func handleIngest(c *gin.Context) {
	s := st()
	if s == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Database unavailable"})
		return
	}
	domainID, err := domain.NormalizeID(domain.IDFromQuery(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	tenantID := tenant.AdminID(c)
	actor := audit.ActorFromContext(c)

	var body ingestRequest
	_ = c.ShouldBindJSON(&body)
	if body.Mode == "" {
		body.Mode = "incremental"
	}
	if body.Source == "" {
		body.Source = "admin"
	}
	sync := body.Sync || strings.EqualFold(c.Query("sync"), "true")

	job, alreadyRunning, err := s.CreateIngestJob(
		c.Request.Context(),
		actor,
		tenantID,
		domainID,
		body.Source,
		body.Mode,
		body.Files,
	)
	if err != nil {
		log.Printf("CreateIngestJob: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !alreadyRunning {
		statsPatch := map[string]any{}
		if id := strings.TrimSpace(body.IndexRunID); id != "" {
			statsPatch["index_run_id"] = id
		}
		if body.ActivateOnComplete {
			statsPatch["activate_on_complete"] = true
		}
		if len(statsPatch) > 0 {
			if err := s.MergeIngestJobStats(c.Request.Context(), job.ID, statsPatch); err != nil {
				log.Printf("MergeIngestJobStats: %v", err)
			} else if refreshed, err := s.GetIngestJob(c.Request.Context(), job.ID); err == nil && refreshed != nil {
				job = *refreshed
			}
		}
		startIngestWorker(job, sync)
		audit.Record(c, audit.Opts{
			Action:   store.AuditActionKBIngest,
			TenantID: tenantID,
			DomainID: domainID,
			Success:  true,
			Details:  map[string]any{"job_id": job.ID, "sync": sync, "mode": body.Mode},
		})
	}
	c.JSON(http.StatusAccepted, gin.H{
		"success":         true,
		"job_id":          job.ID,
		"status":          job.Status,
		"status_label":    ingestStatusLabel(job.Status),
		"already_running": alreadyRunning,
		"sync":            sync,
		"message":         ingestAcceptedMessage(alreadyRunning),
	})
}

func handleIngestStatus(c *gin.Context) {
	s := st()
	if s == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Database unavailable"})
		return
	}
	ctx := c.Request.Context()
	tenantID := tenant.AdminID(c)
	domainID := strings.TrimSpace(c.Query("domain_id"))

	var job *store.IngestJob
	var err error

	if raw := c.Query("job_id"); raw != "" {
		id, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid job_id"})
			return
		}
		job, err = s.GetIngestJob(ctx, id)
	} else if domainID != "" {
		normalized, normErr := domain.NormalizeID(domainID)
		if normErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": normErr.Error()})
			return
		}
		job, err = s.ActiveIngestJob(ctx, tenantID, normalized)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "job_id or domain_id required"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "No ingest job found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"job":          job,
		"status_label": ingestStatusLabel(job.Status),
		"done":         store.IsIngestTerminal(job.Status),
	})
}
