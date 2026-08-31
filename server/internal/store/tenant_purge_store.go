package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

const auditActionTenantPurge = "tenant_purge"

// HasActiveReindexJob reports pending/running reindex for tenant.
func (st *ChatStore) HasActiveReindexJob(ctx context.Context, tenantID string) (bool, error) {
	var n int
	err := st.Pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM reindex_jobs
		WHERE tenant_id = $1 AND status IN ('pending', 'running')`, tenantID,
	).Scan(&n)
	return n > 0, err
}

// PurgeTenant removes all tenant-scoped DB rows, KB registry rows, and upload images.
// Returned storage keys must be deleted from blob storage by the caller.
func (st *ChatStore) PurgeTenant(ctx context.Context, tenantID string) (TenantPurgeStats, []string, error) {
	tenantID = normalizeTenantID(tenantID)
	var stats TenantPurgeStats

	if err := st.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM chat_sessions WHERE tenant_id = $1`, tenantID).Scan(&stats.Sessions); err != nil {
		return stats, nil, err
	}
	if err := st.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM messages m
		JOIN chat_sessions cs ON cs.id = m.session_id
		WHERE cs.tenant_id = $1`, tenantID).Scan(&stats.Messages); err != nil {
		return stats, nil, err
	}
	if err := st.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM message_feedback mf
		JOIN messages m ON m.id = mf.message_id
		JOIN chat_sessions cs ON cs.id = m.session_id
		WHERE cs.tenant_id = $1`, tenantID).Scan(&stats.FeedbackRows); err != nil {
		return stats, nil, err
	}

	rows, err := st.Pool.Query(ctx, `
		SELECT DISTINCT m.image_token FROM messages m
		JOIN chat_sessions cs ON cs.id = m.session_id
		WHERE cs.tenant_id = $1 AND m.image_token IS NOT NULL`, tenantID)
	if err != nil {
		return stats, nil, err
	}
	var tokens []string
	for rows.Next() {
		var tok string
		if err := rows.Scan(&tok); err != nil {
			rows.Close()
			return stats, nil, err
		}
		tokens = append(tokens, tok)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return stats, nil, err
	}
	stats.UploadTokens = int64(len(tokens))

	blobKeys, err := st.ListKBStorageKeysForTenant(ctx, tenantID)
	if err != nil {
		return stats, nil, err
	}
	stats.DataFiles = len(blobKeys)

	tx, err := st.Pool.Begin(ctx)
	if err != nil {
		return stats, nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx, `DELETE FROM chat_sessions WHERE tenant_id = $1`, tenantID)
	if err != nil {
		return stats, nil, fmt.Errorf("delete sessions: %w", err)
	}
	stats.Sessions = tag.RowsAffected()

	tag, err = tx.Exec(ctx, `DELETE FROM analytics_events WHERE payload->>'tenant_id' = $1`, tenantID)
	if err != nil {
		return stats, nil, fmt.Errorf("delete analytics: %w", err)
	}
	stats.AnalyticsRows = tag.RowsAffected()

	tag, err = tx.Exec(ctx, `DELETE FROM reindex_jobs WHERE tenant_id = $1`, tenantID)
	if err != nil {
		return stats, nil, fmt.Errorf("delete reindex jobs: %w", err)
	}
	stats.ReindexJobs = tag.RowsAffected()

	_, err = tx.Exec(ctx, `DELETE FROM ingest_jobs WHERE tenant_id = $1`, tenantID)
	if err != nil {
		return stats, nil, fmt.Errorf("delete ingest jobs: %w", err)
	}

	_, err = tx.Exec(ctx, `DELETE FROM kb_documents WHERE tenant_id = $1`, tenantID)
	if err != nil {
		return stats, nil, fmt.Errorf("delete kb documents: %w", err)
	}

	tag, err = tx.Exec(ctx, `DELETE FROM audit_log WHERE tenant_id = $1 AND action <> $2`, tenantID, auditActionTenantPurge)
	if err != nil {
		return stats, nil, fmt.Errorf("delete audit: %w", err)
	}
	stats.AuditRows = tag.RowsAffected()

	if err := tx.Commit(ctx); err != nil {
		return stats, nil, err
	}

	for _, tok := range tokens {
		_ = os.Remove(filepath.Join(st.uploadDir, tok+".bin"))
	}

	return stats, blobKeys, nil
}
