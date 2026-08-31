package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	IngestStatusQueued    = "queued"
	IngestStatusParsing   = "parsing"
	IngestStatusEmbedding = "embedding"
	IngestStatusIndexing  = "indexing"
	IngestStatusSucceeded = "succeeded"
	IngestStatusFailed    = "failed"
	IngestStatusPartial   = "partial"
)

// IngestJob is one async KB ingestion run.
type IngestJob struct {
	ID           int64    `json:"id"`
	Status       string   `json:"status"`
	TenantID     string   `json:"tenant_id"`
	DomainID     string   `json:"domain_id"`
	Source       string   `json:"source,omitempty"`
	Actor        string   `json:"actor,omitempty"`
	Mode         string   `json:"mode,omitempty"`
	Files        []string `json:"files,omitempty"`
	Stats        any      `json:"stats,omitempty"`
	ErrorMsg     string   `json:"error_msg,omitempty"`
	AttemptCount int      `json:"attempt_count"`
	StartedAt    string   `json:"started_at,omitempty"`
	FinishedAt   string   `json:"finished_at,omitempty"`
	CreatedAt    string   `json:"created_at"`
}

func scanIngestJob(row pgx.Row) (IngestJob, error) {
	var j IngestJob
	var actor, source, mode, errMsg *string
	var filesJSON, statsJSON []byte
	var startedAt, finishedAt *time.Time
	var createdAt time.Time
	err := row.Scan(
		&j.ID, &j.Status, &j.TenantID, &j.DomainID, &source, &actor, &mode,
		&filesJSON, &statsJSON, &errMsg, &j.AttemptCount, &startedAt, &finishedAt, &createdAt,
	)
	if err != nil {
		return j, err
	}
	j.Source = derefString(source)
	j.Actor = derefString(actor)
	j.Mode = derefString(mode)
	j.ErrorMsg = derefString(errMsg)
	j.StartedAt = formatJobTime(startedAt)
	j.FinishedAt = formatJobTime(finishedAt)
	j.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	if len(filesJSON) > 0 {
		_ = json.Unmarshal(filesJSON, &j.Files)
	}
	if len(statsJSON) > 0 {
		var stats any
		if json.Unmarshal(statsJSON, &stats) == nil {
			j.Stats = stats
		}
	}
	return j, nil
}

// CreateIngestJob inserts a queued job or returns the active one for tenant+domain.
func (st *ChatStore) CreateIngestJob(
	ctx context.Context,
	actor, tenantID, domainID, source, mode string,
	files []string,
) (IngestJob, bool, error) {
	var zero IngestJob
	if mode == "" {
		mode = "incremental"
	}
	if source == "" {
		source = "admin"
	}
	filesJSON, err := json.Marshal(files)
	if err != nil {
		return zero, false, err
	}
	row := st.Pool.QueryRow(ctx, `
		INSERT INTO ingest_jobs (status, tenant_id, domain_id, source, actor, mode, files)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)
		RETURNING id, status, tenant_id, domain_id, source, actor, mode,
		          files, stats, error_msg, attempt_count, started_at, finished_at, created_at`,
		IngestStatusQueued, tenantID, domainID, source, nullIfEmpty(actor), mode, filesJSON,
	)
	job, err := scanIngestJob(row)
	if err == nil {
		return job, false, nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		active, err := st.ActiveIngestJob(ctx, tenantID, domainID)
		if err != nil {
			return zero, false, err
		}
		if active == nil {
			return zero, false, fmt.Errorf("ingest job conflict without active row")
		}
		return *active, true, nil
	}
	return zero, false, err
}

// ActiveIngestJob returns pending/running ingest for tenant+domain.
func (st *ChatStore) ActiveIngestJob(ctx context.Context, tenantID, domainID string) (*IngestJob, error) {
	row := st.Pool.QueryRow(ctx, `
		SELECT id, status, tenant_id, domain_id, source, actor, mode,
		       files, stats, error_msg, attempt_count, started_at, finished_at, created_at
		FROM ingest_jobs
		WHERE tenant_id = $1 AND domain_id = $2
		  AND status NOT IN ($3, $4, $5)
		ORDER BY created_at DESC
		LIMIT 1`,
		tenantID, domainID,
		IngestStatusSucceeded, IngestStatusFailed, IngestStatusPartial,
	)
	job, err := scanIngestJob(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// GetIngestJob loads a job by id.
func (st *ChatStore) GetIngestJob(ctx context.Context, id int64) (*IngestJob, error) {
	row := st.Pool.QueryRow(ctx, `
		SELECT id, status, tenant_id, domain_id, source, actor, mode,
		       files, stats, error_msg, attempt_count, started_at, finished_at, created_at
		FROM ingest_jobs WHERE id = $1`, id)
	job, err := scanIngestJob(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// HasActiveIngestJob reports non-terminal ingest for tenant (any domain).
func (st *ChatStore) HasActiveIngestJob(ctx context.Context, tenantID string) (bool, error) {
	var n int
	err := st.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM ingest_jobs
		WHERE tenant_id = $1
		  AND status NOT IN ($2, $3, $4)`,
		tenantID, IngestStatusSucceeded, IngestStatusFailed, IngestStatusPartial,
	).Scan(&n)
	return n > 0, err
}

// MergeIngestJobStats patches ingest_jobs.stats JSON (creates object when null).
func (st *ChatStore) MergeIngestJobStats(ctx context.Context, jobID int64, patch map[string]any) error {
	if len(patch) == 0 {
		return nil
	}
	raw, err := json.Marshal(patch)
	if err != nil {
		return err
	}
	_, err = st.Pool.Exec(ctx, `
		UPDATE ingest_jobs
		SET stats = COALESCE(stats, '{}'::jsonb) || $2::jsonb
		WHERE id = $1`,
		jobID, raw,
	)
	return err
}

// FinishIngestJob marks a job terminal with optional error.
func (st *ChatStore) FinishIngestJob(ctx context.Context, id int64, status, errMsg string) error {
	_, err := st.Pool.Exec(ctx, `
		UPDATE ingest_jobs
		SET status = $1, error_msg = $2, finished_at = NOW()
		WHERE id = $3`,
		status, nullIfEmpty(errMsg), id,
	)
	return err
}

// IsIngestTerminal reports whether a job has finished.
func IsIngestTerminal(status string) bool {
	return status == IngestStatusSucceeded || status == IngestStatusFailed || status == IngestStatusPartial
}
