package admin

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"grounded_llm_server/internal/audit"
	"grounded_llm_server/internal/store"
)

func startReindexWorker(job store.ReindexJob) {
	go func() {
		ctx := context.Background()
		s := st()
		if s == nil {
			return
		}
		if err := s.MarkReindexJobRunning(ctx, job.ID); err != nil {
			log.Printf("reindex job %d mark running: %v", job.ID, err)
		}
		err := triggerRAGReindex()
		if err != nil {
			log.Printf("reindex job %d failed: %v", job.ID, err)
			_ = s.FinishReindexJob(ctx, job.ID, store.ReindexStatusFailed, err.Error())
			recordReindexAuditComplete(job, false, err.Error())
			return
		}
		_ = s.FinishReindexJob(ctx, job.ID, store.ReindexStatusSucceeded, "")
		recordReindexAuditComplete(job, true, "")
	}()
}

func recordReindexAuditComplete(job store.ReindexJob, success bool, errMsg string) {
	details := map[string]any{"job_id": job.ID}
	if errMsg != "" {
		details["error"] = errMsg
	}
	audit.RecordBackground(store.AuditRecord{
		Action:   store.AuditActionKBReindex,
		Actor:    job.Actor,
		TenantID: job.TenantID,
		Success:  success,
		Details:  details,
	})
}

func triggerRAGReindex() error {
	c := cfg()
	if c == nil || c.AdminSecret == "" {
		return fmt.Errorf("ADMIN_SECRET is not set")
	}
	url := strings.TrimRight(c.PythonBaseURL, "/") + "/admin/reindex"
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Admin-Secret", c.AdminSecret)
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("python reindex HTTP %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func isReindexTerminal(status string) bool {
	return status == store.ReindexStatusSucceeded || status == store.ReindexStatusFailed
}

func reindexStatusLabel(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case store.ReindexStatusPending:
		return "queued"
	case store.ReindexStatusRunning:
		return "running"
	case store.ReindexStatusSucceeded:
		return "succeeded"
	case store.ReindexStatusFailed:
		return "failed"
	default:
		return status
	}
}

func reindexAcceptedMessage(alreadyRunning bool) string {
	if alreadyRunning {
		return "RAG reindex already in progress"
	}
	return "RAG reindex queued"
}

// IsReindexTerminal reports whether a reindex job has finished.
func IsReindexTerminal(status string) bool {
	return isReindexTerminal(status)
}

// ReindexStatusLabel returns a human-readable reindex status label.
func ReindexStatusLabel(status string) string {
	return reindexStatusLabel(status)
}

// ReindexAcceptedMessage returns the HTTP message for a reindex enqueue response.
func ReindexAcceptedMessage(alreadyRunning bool) string {
	return reindexAcceptedMessage(alreadyRunning)
}
