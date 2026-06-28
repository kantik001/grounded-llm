-- Async RAG reindex jobs (admin panel)

CREATE TABLE IF NOT EXISTS reindex_jobs (
    id          BIGSERIAL PRIMARY KEY,
    status      TEXT NOT NULL CHECK (status IN ('pending', 'running', 'succeeded', 'failed')),
    actor       TEXT,
    tenant_id   TEXT,
    domain_id   TEXT,
    error_msg   TEXT,
    started_at  TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reindex_jobs_created ON reindex_jobs (created_at DESC);

-- At most one pending/running job at a time (global Chroma rebuild).
CREATE UNIQUE INDEX IF NOT EXISTS idx_reindex_jobs_one_active
    ON reindex_jobs ((true))
    WHERE status IN ('pending', 'running');
