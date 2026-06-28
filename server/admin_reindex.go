package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// POST /admin/reindex — enqueue async RAG reindex (returns job id immediately).
func handleAdminReindex(c *gin.Context) {
	if chatStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Database unavailable"})
		return
	}
	actor := adminActorFromContext(c)
	tenantID := adminTenantID(c)
	domainID := c.Query("domain_id")
	if domainID == "" {
		domainID = c.PostForm("domain_id")
	}

	job, alreadyRunning, err := chatStore.CreateReindexJob(c.Request.Context(), actor, tenantID, domainID)
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

func reindexAcceptedMessage(alreadyRunning bool) string {
	if alreadyRunning {
		return "RAG reindex already in progress"
	}
	return "RAG reindex queued"
}

// GET /admin/reindex/status?job_id= — poll job status (latest active job if job_id omitted).
func handleAdminReindexStatus(c *gin.Context) {
	if chatStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Database unavailable"})
		return
	}
	ctx := c.Request.Context()
	var job *ReindexJob
	var err error

	if raw := c.Query("job_id"); raw != "" {
		id, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid job_id"})
			return
		}
		job, err = chatStore.GetReindexJob(ctx, id)
	} else {
		job, err = chatStore.ActiveReindexJob(ctx)
		if err == nil && job == nil {
			job, err = chatStore.GetLatestReindexJob(ctx)
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
