package store

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
)

const (
	KBOutboxStatusPending    = "pending"
	KBOutboxStatusProcessing = "processing"
	KBOutboxStatusDone       = "done"
	KBOutboxStatusFailed     = "failed"
)

// KBIngestOutboxRow is one pending registry change awaiting ingest enqueue.
type KBIngestOutboxRow struct {
	ID            int64
	TenantID      string
	DomainID      string
	DocumentID    string
	VersionID     string
	LogicalKey    string
	ContentSHA256 string
	Source        string
}

// KBAutoIngestEnabled reports whether registry upserts should enqueue ingest outbox rows.
func KBAutoIngestEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("KB_AUTO_INGEST")))
	return v == "1" || v == "true" || v == "yes"
}

// EnqueueKBIngestOutboxTx inserts an outbox row in the caller transaction.
func EnqueueKBIngestOutboxTx(
	ctx context.Context,
	tx pgx.Tx,
	tenantID, domainID, documentID, versionID, logicalKey, contentSHA256, source string,
) error {
	if !KBAutoIngestEnabled() {
		return nil
	}
	if source == "" {
		source = "upload"
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO kb_ingest_outbox
			(tenant_id, domain_id, document_id, version_id, logical_key, content_sha256, source, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		tenantID, domainID, documentID, versionID, logicalKey, contentSHA256, source, KBOutboxStatusPending,
	)
	return err
}

// ClaimPendingKBIngestOutbox locks pending rows for tenant+domain and marks them processing.
func (st *ChatStore) ClaimPendingKBIngestOutbox(ctx context.Context, tenantID, domainID string) ([]KBIngestOutboxRow, error) {
	tx, err := st.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, `
		SELECT id, tenant_id, domain_id, document_id, version_id, logical_key, content_sha256, source
		FROM kb_ingest_outbox
		WHERE status = $1 AND tenant_id = $2 AND domain_id = $3
		ORDER BY created_at
		FOR UPDATE SKIP LOCKED`,
		KBOutboxStatusPending, tenantID, domainID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []KBIngestOutboxRow
	var ids []int64
	for rows.Next() {
		var row KBIngestOutboxRow
		if err := rows.Scan(
			&row.ID, &row.TenantID, &row.DomainID, &row.DocumentID, &row.VersionID,
			&row.LogicalKey, &row.ContentSHA256, &row.Source,
		); err != nil {
			return nil, err
		}
		out = append(out, row)
		ids = append(ids, row.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return nil, nil
	}

	if _, err := tx.Exec(ctx, `
		UPDATE kb_ingest_outbox SET status = $1 WHERE id = ANY($2)`,
		KBOutboxStatusProcessing, ids,
	); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return out, nil
}

// FinishKBIngestOutbox marks outbox rows done or failed after ingest enqueue.
func (st *ChatStore) FinishKBIngestOutbox(ctx context.Context, ids []int64, jobID int64, errMsg string) error {
	if len(ids) == 0 {
		return nil
	}
	status := KBOutboxStatusDone
	var jobPtr *int64
	if errMsg != "" {
		status = KBOutboxStatusFailed
	} else if jobID > 0 {
		jobPtr = &jobID
	}
	_, err := st.Pool.Exec(ctx, `
		UPDATE kb_ingest_outbox
		SET status = $1,
		    ingest_job_id = COALESCE($2, ingest_job_id),
		    error_msg = NULLIF($3, ''),
		    processed_at = NOW()
		WHERE id = ANY($4)`,
		status, jobPtr, errMsg, ids,
	)
	return err
}

// FlushKBIngestOutbox claims pending rows and creates an ingest job for the scope.
func (st *ChatStore) FlushKBIngestOutbox(
	ctx context.Context,
	tenantID, domainID, actor, source string,
	files []string,
) (IngestJob, bool, int, error) {
	var zero IngestJob
	rows, err := st.ClaimPendingKBIngestOutbox(ctx, tenantID, domainID)
	if err != nil {
		return zero, false, 0, err
	}
	if len(rows) == 0 {
		return zero, false, 0, nil
	}

	ids := make([]int64, 0, len(rows))
	keys := make([]string, 0, len(rows))
	seen := map[string]bool{}
	if source == "" {
		source = rows[0].Source
	}
	for _, row := range rows {
		ids = append(ids, row.ID)
		if row.LogicalKey != "" && !seen[row.LogicalKey] {
			seen[row.LogicalKey] = true
			keys = append(keys, row.LogicalKey)
		}
	}
	targetFiles := files
	if len(targetFiles) == 0 {
		targetFiles = keys
	}

	job, alreadyRunning, err := st.CreateIngestJob(ctx, actor, tenantID, domainID, source, "incremental", targetFiles)
	if err != nil {
		_ = st.FinishKBIngestOutbox(ctx, ids, 0, err.Error())
		return zero, false, len(rows), fmt.Errorf("create ingest job: %w", err)
	}
	if err := st.FinishKBIngestOutbox(ctx, ids, job.ID, ""); err != nil {
		return job, alreadyRunning, len(rows), err
	}
	return job, alreadyRunning, len(rows), nil
}
