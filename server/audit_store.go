package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	auditActionLoginFailed = "admin_login_failed"
	auditActionLogin       = "admin_login"
	auditActionKBUpload    = "kb_upload"
	auditActionKBDelete    = "kb_delete"
	auditActionKBReindex   = "kb_reindex"
)

const (
	auditLogDefaultLimit = 50
	auditLogMaxLimit     = 200
)

// AuditEntry is one row returned from GET /admin/audit-log.
type AuditEntry struct {
	ID         int64          `json:"id"`
	OccurredAt string         `json:"occurred_at"`
	Action     string         `json:"action"`
	Actor      string         `json:"actor,omitempty"`
	TenantID   string         `json:"tenant_id,omitempty"`
	DomainID   string         `json:"domain_id,omitempty"`
	Resource   string         `json:"resource,omitempty"`
	ClientIP   string         `json:"client_ip,omitempty"`
	RequestID  string         `json:"request_id,omitempty"`
	Success    bool           `json:"success"`
	Details    map[string]any `json:"details,omitempty"`
}

type auditRecord struct {
	Action    string
	Actor     string
	TenantID  string
	DomainID  string
	Resource  string
	ClientIP  string
	RequestID string
	Success   bool
	Details   map[string]any
}

// RecordAudit persists an admin audit event.
func (st *ChatStore) RecordAudit(ctx context.Context, rec auditRecord) error {
	if st == nil || st.pool == nil {
		return fmt.Errorf("chat store not initialized")
	}
	details := rec.Details
	if details == nil {
		details = map[string]any{}
	}
	raw, err := json.Marshal(details)
	if err != nil {
		return err
	}
	_, err = st.pool.Exec(ctx, `
		INSERT INTO audit_log (
			action, actor, tenant_id, domain_id, resource,
			client_ip, request_id, success, details
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		rec.Action,
		nullIfEmpty(rec.Actor),
		nullIfEmpty(rec.TenantID),
		nullIfEmpty(rec.DomainID),
		nullIfEmpty(rec.Resource),
		nullIfEmpty(rec.ClientIP),
		nullIfEmpty(rec.RequestID),
		rec.Success,
		raw,
	)
	return err
}

// ListAuditLog returns recent audit events (newest first).
func (st *ChatStore) ListAuditLog(ctx context.Context, limit, offset int, actionFilter string) ([]AuditEntry, error) {
	if st == nil || st.pool == nil {
		return nil, fmt.Errorf("chat store not initialized")
	}
	if limit <= 0 {
		limit = auditLogDefaultLimit
	}
	if limit > auditLogMaxLimit {
		limit = auditLogMaxLimit
	}
	if offset < 0 {
		offset = 0
	}

	actionFilter = strings.TrimSpace(actionFilter)
	var rows pgx.Rows
	var err error
	if actionFilter != "" {
		rows, err = st.pool.Query(ctx, `
			SELECT id, occurred_at, action, COALESCE(actor, ''), COALESCE(tenant_id, ''),
			       COALESCE(domain_id, ''), COALESCE(resource, ''), COALESCE(client_ip, ''),
			       COALESCE(request_id, ''), success, details
			FROM audit_log
			WHERE action = $1
			ORDER BY occurred_at DESC, id DESC
			LIMIT $2 OFFSET $3`, actionFilter, limit, offset)
	} else {
		rows, err = st.pool.Query(ctx, `
			SELECT id, occurred_at, action, COALESCE(actor, ''), COALESCE(tenant_id, ''),
			       COALESCE(domain_id, ''), COALESCE(resource, ''), COALESCE(client_ip, ''),
			       COALESCE(request_id, ''), success, details
			FROM audit_log
			ORDER BY occurred_at DESC, id DESC
			LIMIT $1 OFFSET $2`, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AuditEntry
	for rows.Next() {
		var e AuditEntry
		var occurred time.Time
		var detailsJSON []byte
		if err := rows.Scan(
			&e.ID, &occurred, &e.Action, &e.Actor, &e.TenantID, &e.DomainID,
			&e.Resource, &e.ClientIP, &e.RequestID, &e.Success, &detailsJSON,
		); err != nil {
			return nil, err
		}
		e.OccurredAt = occurred.UTC().Format(time.RFC3339)
		if len(detailsJSON) > 0 {
			_ = json.Unmarshal(detailsJSON, &e.Details)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
