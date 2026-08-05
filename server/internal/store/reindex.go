package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	ReindexStatusPending   = "pending"
	ReindexStatusRunning   = "running"
	ReindexStatusSucceeded = "succeeded"
	ReindexStatusFailed    = "failed"
)

// ReindexJob is one async RAG reindex run.
type ReindexJob struct {
	ID         int64  `json:"id"`
	Status     string `json:"status"`
	Actor      string `json:"actor,omitempty"`
	TenantID   string `json:"tenant_id,omitempty"`
	DomainID   string `json:"domain_id,omitempty"`
	ErrorMsg   string `json:"error_msg,omitempty"`
	StartedAt  string `json:"started_at,omitempty"`
	FinishedAt string `json:"finished_at,omitempty"`
	CreatedAt  string `json:"created_at"`
}

func formatJobTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func scanReindexJob(row pgx.Row) (ReindexJob, error) {
	var j ReindexJob
	var actor, tenantID, domainID, errMsg *string
	var startedAt, finishedAt *time.Time
	var createdAt time.Time
	err := row.Scan(&j.ID, &j.Status, &actor, &tenantID, &domainID, &errMsg, &startedAt, &finishedAt, &createdAt)
	if err != nil {
		return j, err
	}
	j.Actor = derefString(actor)
	j.TenantID = derefString(tenantID)
	j.DomainID = derefString(domainID)
	j.ErrorMsg = derefString(errMsg)
	j.StartedAt = formatJobTime(startedAt)
	j.FinishedAt = formatJobTime(finishedAt)
	j.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	return j, nil
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// CreateReindexJob inserts a pending job or returns the active one if already running.
func (st *ChatStore) CreateReindexJob(ctx context.Context, actor, tenantID, domainID string) (ReindexJob, bool, error) {
	var zero ReindexJob
	row := st.Pool.QueryRow(ctx, `
		INSERT INTO reindex_jobs (status, actor, tenant_id, domain_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, status, actor, tenant_id, domain_id, error_msg, started_at, finished_at, created_at`,
		ReindexStatusPending, nullIfEmpty(actor), nullIfEmpty(tenantID), nullIfEmpty(domainID),
	)
	job, err := scanReindexJob(row)
	if err == nil {
		return job, false, nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		active, err := st.ActiveReindexJob(ctx)
		if err != nil {
			return zero, false, err
		}
		if active == nil {
			return zero, false, fmt.Errorf("reindex job conflict without active row")
		}
		return *active, true, nil
	}
	return zero, false, err
}

// ActiveReindexJob returns the current pending/running job, if any.
func (st *ChatStore) ActiveReindexJob(ctx context.Context) (*ReindexJob, error) {
	row := st.Pool.QueryRow(ctx, `
		SELECT id, status, actor, tenant_id, domain_id, error_msg, started_at, finished_at, created_at
		FROM reindex_jobs
		WHERE status IN ($1, $2)
		ORDER BY created_at DESC
		LIMIT 1`,
		ReindexStatusPending, ReindexStatusRunning,
	)
	job, err := scanReindexJob(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// GetReindexJob loads a job by id.
func (st *ChatStore) GetReindexJob(ctx context.Context, id int64) (*ReindexJob, error) {
	row := st.Pool.QueryRow(ctx, `
		SELECT id, status, actor, tenant_id, domain_id, error_msg, started_at, finished_at, created_at
		FROM reindex_jobs WHERE id = $1`, id)
	job, err := scanReindexJob(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// MarkReindexJobRunning sets a pending job to running.
func (st *ChatStore) MarkReindexJobRunning(ctx context.Context, id int64) error {
	_, err := st.Pool.Exec(ctx, `
		UPDATE reindex_jobs
		SET status = $1, started_at = COALESCE(started_at, NOW())
		WHERE id = $2 AND status = $3`,
		ReindexStatusRunning, id, ReindexStatusPending,
	)
	return err
}

// FinishReindexJob marks a job terminal with optional error.
func (st *ChatStore) FinishReindexJob(ctx context.Context, id int64, status, errMsg string) error {
	_, err := st.Pool.Exec(ctx, `
		UPDATE reindex_jobs
		SET status = $1, error_msg = $2, finished_at = NOW()
		WHERE id = $3`,
		status, nullIfEmpty(errMsg), id,
	)
	return err
}

// GetLatestReindexJob returns the most recent job regardless of status.
func (st *ChatStore) GetLatestReindexJob(ctx context.Context) (*ReindexJob, error) {
	row := st.Pool.QueryRow(ctx, `
		SELECT id, status, actor, tenant_id, domain_id, error_msg, started_at, finished_at, created_at
		FROM reindex_jobs
		ORDER BY created_at DESC
		LIMIT 1`)
	job, err := scanReindexJob(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}
