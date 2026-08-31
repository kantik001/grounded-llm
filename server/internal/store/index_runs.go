package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	IndexRunStatusBuilding = "building"
	IndexRunStatusActive   = "active"
	IndexRunStatusRetired  = "retired"
	IndexRunStatusFailed   = "failed"
)

// IndexRun tracks one disposable index generation.
type IndexRun struct {
	ID             string         `json:"id"`
	TenantID       string         `json:"tenant_id,omitempty"`
	DomainID       string         `json:"domain_id,omitempty"`
	Backend        string         `json:"backend"`
	EmbeddingModel string         `json:"embedding_model"`
	ChunkSchema    map[string]any `json:"chunk_schema,omitempty"`
	Status         string         `json:"status"`
	ErrorMsg       string         `json:"error_msg,omitempty"`
	CreatedAt      string         `json:"created_at"`
	ActivatedAt    string         `json:"activated_at,omitempty"`
}

func scanIndexRun(row pgx.Row) (IndexRun, error) {
	var run IndexRun
	var tenantID, domainID, errMsg *string
	var chunkJSON []byte
	var createdAt time.Time
	var activatedAt *time.Time
	err := row.Scan(
		&run.ID, &tenantID, &domainID, &run.Backend, &run.EmbeddingModel, &chunkJSON,
		&run.Status, &errMsg, &createdAt, &activatedAt,
	)
	if err != nil {
		return run, err
	}
	if tenantID != nil {
		run.TenantID = *tenantID
	}
	if domainID != nil {
		run.DomainID = *domainID
	}
	if errMsg != nil {
		run.ErrorMsg = *errMsg
	}
	if len(chunkJSON) > 0 {
		_ = json.Unmarshal(chunkJSON, &run.ChunkSchema)
	}
	run.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	if activatedAt != nil {
		run.ActivatedAt = activatedAt.UTC().Format(time.RFC3339)
	}
	return run, nil
}

// CreateIndexRun inserts a building index run.
func (st *ChatStore) CreateIndexRun(
	ctx context.Context,
	tenantID, domainID, backend, embeddingModel string,
	chunkSchema map[string]any,
) (IndexRun, error) {
	var zero IndexRun
	if backend == "" {
		backend = "chroma"
	}
	chunkJSON, err := json.Marshal(chunkSchema)
	if err != nil {
		return zero, err
	}
	id := uuid.NewString()
	var tPtr, dPtr *string
	if tenantID != "" {
		tPtr = &tenantID
	}
	if domainID != "" {
		dPtr = &domainID
	}
	row := st.Pool.QueryRow(ctx, `
		INSERT INTO index_runs (id, tenant_id, domain_id, backend, embedding_model, chunk_schema, status)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7)
		RETURNING id, tenant_id, domain_id, backend, embedding_model, chunk_schema, status, error_msg, created_at, activated_at`,
		id, tPtr, dPtr, backend, embeddingModel, chunkJSON, IndexRunStatusBuilding,
	)
	return scanIndexRun(row)
}

// ActivateIndexRun marks a run active and retires the previous active run for scope.
func (st *ChatStore) ActivateIndexRun(ctx context.Context, tenantID, domainID, runID string) error {
	tx, err := st.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `
		UPDATE index_runs SET status = $1
		WHERE tenant_id IS NOT DISTINCT FROM NULLIF($2, '')::text
		  AND domain_id IS NOT DISTINCT FROM NULLIF($3, '')::text
		  AND status = $4 AND id <> $5`,
		IndexRunStatusRetired, tenantID, domainID, IndexRunStatusActive, runID,
	)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		UPDATE index_runs SET status = $1, activated_at = NOW() WHERE id = $2`,
		IndexRunStatusActive, runID,
	)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO index_run_active (tenant_id, domain_id, index_run_id, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (tenant_id, domain_id) DO UPDATE SET index_run_id = EXCLUDED.index_run_id, updated_at = NOW()`,
		tenantID, domainID, runID,
	)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ActiveIndexRunID returns the active index run for tenant+domain scope.
func (st *ChatStore) ActiveIndexRunID(ctx context.Context, tenantID, domainID string) (string, error) {
	var runID string
	err := st.Pool.QueryRow(ctx, `
		SELECT index_run_id FROM index_run_active WHERE tenant_id = $1 AND domain_id = $2`,
		tenantID, domainID,
	).Scan(&runID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return runID, err
}

// UpsertIndexDocumentState records indexing outcome for a document in a run.
func (st *ChatStore) UpsertIndexDocumentState(
	ctx context.Context,
	runID, documentID string,
	indexedVersion int,
	contentSHA256 string,
	chunkCount int,
) error {
	_, err := st.Pool.Exec(ctx, `
		INSERT INTO index_document_state (index_run_id, document_id, indexed_version, content_sha256, chunk_count, indexed_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (index_run_id, document_id) DO UPDATE SET
			indexed_version = EXCLUDED.indexed_version,
			content_sha256 = EXCLUDED.content_sha256,
			chunk_count = EXCLUDED.chunk_count,
			indexed_at = NOW()`,
		runID, documentID, indexedVersion, contentSHA256, chunkCount,
	)
	return err
}

// GetIndexRun loads an index run by id.
func (st *ChatStore) GetIndexRun(ctx context.Context, runID string) (*IndexRun, error) {
	row := st.Pool.QueryRow(ctx, `
		SELECT id, tenant_id, domain_id, backend, embedding_model, chunk_schema, status, error_msg, created_at, activated_at
		FROM index_runs WHERE id = $1`, runID)
	run, err := scanIndexRun(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &run, nil
}

// FinishIndexRun marks a run failed with message.
func (st *ChatStore) FinishIndexRun(ctx context.Context, runID, status, errMsg string) error {
	if status == "" {
		status = IndexRunStatusFailed
	}
	_, err := st.Pool.Exec(ctx, `
		UPDATE index_runs SET status = $1, error_msg = $2 WHERE id = $3`,
		status, nullIfEmpty(errMsg), runID,
	)
	return err
}

// DefaultChunkSchema returns the current chunking profile for index runs.
func DefaultChunkSchema() map[string]any {
	return map[string]any{
		"chunk_size":    500,
		"chunk_overlap": 50,
		"splitter":      "recursive",
		"schema":        1,
	}
}

// EnsureActiveIndexRun creates and activates an index run if none exists for scope.
func (st *ChatStore) EnsureActiveIndexRun(ctx context.Context, tenantID, domainID, backend, embeddingModel string) (string, error) {
	runID, err := st.ActiveIndexRunID(ctx, tenantID, domainID)
	if err != nil {
		return "", err
	}
	if runID != "" {
		return runID, nil
	}
	run, err := st.CreateIndexRun(ctx, tenantID, domainID, backend, embeddingModel, DefaultChunkSchema())
	if err != nil {
		return "", err
	}
	if err := st.ActivateIndexRun(ctx, tenantID, domainID, run.ID); err != nil {
		return "", fmt.Errorf("activate index run: %w", err)
	}
	return run.ID, nil
}
