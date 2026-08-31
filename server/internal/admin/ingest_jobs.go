package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"grounded_llm_server/internal/audit"
	"grounded_llm_server/internal/store"
)

func startIngestWorker(job store.IngestJob, sync bool) {
	go func() {
		if err := triggerPythonIngest(job.ID, sync); err != nil {
			log.Printf("ingest job %d trigger failed: %v", job.ID, err)
			s := st()
			if s != nil {
				_ = s.FinishIngestJob(context.Background(), job.ID, store.IngestStatusFailed, err.Error())
			}
			recordIngestAuditComplete(job, false, err.Error())
		}
	}()
}

func recordIngestAuditComplete(job store.IngestJob, success bool, errMsg string) {
	details := map[string]any{"job_id": job.ID, "mode": job.Mode}
	if errMsg != "" {
		details["error"] = errMsg
	}
	audit.RecordBackground(store.AuditRecord{
		Action:   store.AuditActionKBIngest,
		Actor:    job.Actor,
		TenantID: job.TenantID,
		DomainID: job.DomainID,
		Success:  success,
		Details:  details,
	})
}

func triggerPythonIngest(jobID int64, sync bool) error {
	c := cfg()
	if c == nil || c.AdminSecret == "" {
		return fmt.Errorf("ADMIN_SECRET is not set")
	}
	body, _ := json.Marshal(map[string]any{
		"job_id": jobID,
		"sync":   sync,
	})
	url := strings.TrimRight(c.PythonBaseURL, "/") + "/admin/ingest/run"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Secret", c.AdminSecret)
	timeout := 2 * time.Minute
	if sync {
		timeout = 15 * time.Minute
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("python ingest HTTP %d: %s", resp.StatusCode, string(raw))
	}
	return nil
}

func ingestStatusLabel(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case store.IngestStatusQueued:
		return "queued"
	case store.IngestStatusParsing:
		return "parsing"
	case store.IngestStatusEmbedding:
		return "embedding"
	case store.IngestStatusIndexing:
		return "indexing"
	case store.IngestStatusSucceeded:
		return "succeeded"
	case store.IngestStatusFailed:
		return "failed"
	case store.IngestStatusPartial:
		return "partial"
	default:
		return status
	}
}

func ingestAcceptedMessage(alreadyRunning bool) string {
	if alreadyRunning {
		return "KB ingest already in progress for this tenant/domain"
	}
	return "KB ingest queued"
}

// IngestStatusLabel returns a human-readable ingest status label.
func IngestStatusLabel(status string) string {
	return ingestStatusLabel(status)
}

// IngestAcceptedMessage returns the HTTP message for an ingest enqueue response.
func IngestAcceptedMessage(alreadyRunning bool) string {
	return ingestAcceptedMessage(alreadyRunning)
}
