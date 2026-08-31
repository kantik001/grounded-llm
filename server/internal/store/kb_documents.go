package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	KBDocStatusActive   = "active"
	KBDocStatusDeleted  = "deleted"
	KBDocStatusArchived = "archived"
)

// KBDocument is the logical document record in Postgres (SoT metadata).
type KBDocument struct {
	ID             string `json:"id"`
	TenantID       string `json:"tenant_id"`
	DomainID       string `json:"domain_id"`
	LogicalKey     string `json:"logical_key"`
	Title          string `json:"title,omitempty"`
	MimeType       string `json:"mime_type,omitempty"`
	Status         string `json:"status"`
	CurrentVersion int    `json:"current_version"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// KBDocumentVersion is one immutable blob version.
type KBDocumentVersion struct {
	ID            string         `json:"id"`
	DocumentID    string         `json:"document_id"`
	Version       int            `json:"version"`
	StorageKey    string         `json:"storage_key"`
	ContentSHA256 string         `json:"content_sha256"`
	SizeBytes     int64          `json:"size_bytes"`
	Source        string         `json:"source"`
	SourceRef     map[string]any `json:"source_ref,omitempty"`
	CreatedBy     string         `json:"created_by,omitempty"`
	CreatedAt     string         `json:"created_at"`
}

// UpsertKBDocumentInput for upload/sync.
type UpsertKBDocumentInput struct {
	TenantID      string
	DomainID      string
	LogicalKey    string
	Title         string
	MimeType      string
	StorageKey    string
	ContentSHA256 string
	SizeBytes     int64
	Source        string
	SourceRef     map[string]any
	CreatedBy     string
}

func scanKBDocument(row pgx.Row) (KBDocument, error) {
	var doc KBDocument
	var createdAt, updatedAt time.Time
	err := row.Scan(
		&doc.ID, &doc.TenantID, &doc.DomainID, &doc.LogicalKey, &doc.Title, &doc.MimeType,
		&doc.Status, &doc.CurrentVersion, &createdAt, &updatedAt,
	)
	if err != nil {
		return doc, err
	}
	doc.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	doc.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
	return doc, nil
}

func scanKBDocumentVersion(row pgx.Row) (KBDocumentVersion, error) {
	var ver KBDocumentVersion
	var sourceRefJSON []byte
	var createdAt time.Time
	err := row.Scan(
		&ver.ID, &ver.DocumentID, &ver.Version, &ver.StorageKey, &ver.ContentSHA256,
		&ver.SizeBytes, &ver.Source, &sourceRefJSON, &ver.CreatedBy, &createdAt,
	)
	if err != nil {
		return ver, err
	}
	if len(sourceRefJSON) > 0 {
		_ = json.Unmarshal(sourceRefJSON, &ver.SourceRef)
	}
	ver.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	return ver, nil
}

// UpsertKBDocument creates or updates a document and appends a new version.
func (st *ChatStore) UpsertKBDocument(ctx context.Context, in UpsertKBDocumentInput) (KBDocument, KBDocumentVersion, error) {
	var zeroDoc KBDocument
	var zeroVer KBDocumentVersion
	if strings.TrimSpace(in.LogicalKey) == "" {
		return zeroDoc, zeroVer, fmt.Errorf("logical_key required")
	}
	if in.Source == "" {
		in.Source = "upload"
	}
	sourceRefJSON, err := json.Marshal(in.SourceRef)
	if err != nil {
		return zeroDoc, zeroVer, err
	}

	tx, err := st.Pool.Begin(ctx)
	if err != nil {
		return zeroDoc, zeroVer, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var doc KBDocument
	row := tx.QueryRow(ctx, `
		SELECT id, tenant_id, domain_id, logical_key, title, mime_type, status, current_version, created_at, updated_at
		FROM kb_documents
		WHERE tenant_id = $1 AND domain_id = $2 AND logical_key = $3`,
		in.TenantID, in.DomainID, in.LogicalKey,
	)
	doc, err = scanKBDocument(row)
	if errors.Is(err, pgx.ErrNoRows) {
		docID := uuid.NewString()
		row = tx.QueryRow(ctx, `
			INSERT INTO kb_documents (id, tenant_id, domain_id, logical_key, title, mime_type, status, current_version)
			VALUES ($1, $2, $3, $4, $5, $6, $7, 0)
			RETURNING id, tenant_id, domain_id, logical_key, title, mime_type, status, current_version, created_at, updated_at`,
			docID, in.TenantID, in.DomainID, in.LogicalKey, in.Title, in.MimeType, KBDocStatusActive,
		)
		doc, err = scanKBDocument(row)
	}
	if err != nil {
		return zeroDoc, zeroVer, err
	}

	nextVersion := doc.CurrentVersion + 1
	if in.StorageKey == "" {
		ext := strings.TrimPrefix(filepath.Ext(in.LogicalKey), ".")
		if ext != "" {
			ext = "." + ext
		}
		in.StorageKey = fmt.Sprintf(
			"tenants/%s/domains/%s/docs/%s/v%d/%s%s",
			in.TenantID, in.DomainID, doc.ID, nextVersion, in.ContentSHA256, ext,
		)
	}
	verID := uuid.NewString()
	row = tx.QueryRow(ctx, `
		INSERT INTO kb_document_versions
			(id, document_id, version, storage_key, content_sha256, size_bytes, source, source_ref, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9)
		RETURNING id, document_id, version, storage_key, content_sha256, size_bytes, source, source_ref, created_by, created_at`,
		verID, doc.ID, nextVersion, in.StorageKey, in.ContentSHA256, in.SizeBytes, in.Source, sourceRefJSON, in.CreatedBy,
	)
	ver, err := scanKBDocumentVersion(row)
	if err != nil {
		return zeroDoc, zeroVer, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE kb_documents
		SET current_version = $1, title = COALESCE(NULLIF($2, ''), title),
		    mime_type = COALESCE(NULLIF($3, ''), mime_type),
		    status = $4, updated_at = NOW()
		WHERE id = $5`,
		nextVersion, in.Title, in.MimeType, KBDocStatusActive, doc.ID,
	)
	if err != nil {
		return zeroDoc, zeroVer, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO kb_document_acl (document_id, principal_type, principal_id, permission)
		VALUES ($1, 'tenant', $2, 'read')
		ON CONFLICT (document_id, principal_type, principal_id) DO NOTHING`,
		doc.ID, in.TenantID,
	)
	if err != nil {
		return zeroDoc, zeroVer, err
	}

	if err := tx.Commit(ctx); err != nil {
		return zeroDoc, zeroVer, err
	}
	doc.CurrentVersion = nextVersion
	return doc, ver, nil
}

// ListKBDocuments returns active documents for tenant+domain.
func (st *ChatStore) ListKBDocuments(ctx context.Context, tenantID, domainID string) ([]KBDocument, error) {
	rows, err := st.Pool.Query(ctx, `
		SELECT id, tenant_id, domain_id, logical_key, title, mime_type, status, current_version, created_at, updated_at
		FROM kb_documents
		WHERE tenant_id = $1 AND domain_id = $2 AND status = $3
		ORDER BY logical_key`,
		tenantID, domainID, KBDocStatusActive,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []KBDocument
	for rows.Next() {
		doc, err := scanKBDocument(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, doc)
	}
	return out, rows.Err()
}

// GetKBDocumentVersion loads a specific version by id.
func (st *ChatStore) GetKBDocumentVersion(ctx context.Context, versionID string) (*KBDocumentVersion, error) {
	row := st.Pool.QueryRow(ctx, `
		SELECT id, document_id, version, storage_key, content_sha256, size_bytes, source, source_ref, created_by, created_at
		FROM kb_document_versions WHERE id = $1`, versionID)
	ver, err := scanKBDocumentVersion(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ver, nil
}

// MarkKBDocumentDeleted soft-deletes a document by logical key.
func (st *ChatStore) MarkKBDocumentDeleted(ctx context.Context, tenantID, domainID, logicalKey string) error {
	_, err := st.Pool.Exec(ctx, `
		UPDATE kb_documents SET status = $1, updated_at = NOW()
		WHERE tenant_id = $2 AND domain_id = $3 AND logical_key = $4`,
		KBDocStatusDeleted, tenantID, domainID, logicalKey,
	)
	return err
}

// AllowedDocumentIDsForPrincipal returns document ids readable by principal (tenant-wide + role/user/group).
func (st *ChatStore) AllowedDocumentIDsForPrincipal(
	ctx context.Context,
	tenantID string,
	principals []struct{ Type, ID string },
) ([]string, error) {
	if len(principals) == 0 {
		principals = []struct{ Type, ID string }{{Type: "tenant", ID: tenantID}}
	}
	ids := make([]string, 0, 32)
	seen := map[string]bool{}
	for _, p := range principals {
		rows, err := st.Pool.Query(ctx, `
			SELECT DISTINCT d.id
			FROM kb_documents d
			JOIN kb_document_acl a ON a.document_id = d.id
			WHERE d.tenant_id = $1 AND d.status = $2
			  AND a.principal_type = $3 AND a.principal_id = $4
			  AND a.permission IN ('read', 'admin')`,
			tenantID, KBDocStatusActive, p.Type, p.ID,
		)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return nil, err
			}
			if !seen[id] {
				seen[id] = true
				ids = append(ids, id)
			}
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}
	return ids, nil
}
