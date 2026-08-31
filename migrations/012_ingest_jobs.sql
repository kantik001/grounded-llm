-- Async KB ingestion jobs (parse → embed → index pipeline)

CREATE TABLE IF NOT EXISTS ingest_jobs (
    id           BIGSERIAL PRIMARY KEY,
    status       TEXT NOT NULL CHECK (status IN (
        'queued', 'parsing', 'embedding', 'indexing', 'succeeded', 'failed', 'partial'
    )),
    tenant_id    TEXT NOT NULL,
    domain_id    TEXT NOT NULL,
    source       TEXT NOT NULL DEFAULT 'admin',
    actor        TEXT,
    mode         TEXT NOT NULL DEFAULT 'incremental' CHECK (mode IN ('incremental', 'full')),
    files        JSONB NOT NULL DEFAULT '[]',
    stats        JSONB NOT NULL DEFAULT '{}',
    error_msg    TEXT,
    attempt_count INT NOT NULL DEFAULT 0,
    started_at   TIMESTAMPTZ,
    finished_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ingest_jobs_created ON ingest_jobs (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ingest_jobs_tenant_domain ON ingest_jobs (tenant_id, domain_id, created_at DESC);

-- One active ingest per tenant+domain scope.
CREATE UNIQUE INDEX IF NOT EXISTS idx_ingest_jobs_one_active_scope
    ON ingest_jobs (tenant_id, domain_id)
    WHERE status NOT IN ('succeeded', 'failed', 'partial');

CREATE TABLE IF NOT EXISTS ingest_tasks (
    id            BIGSERIAL PRIMARY KEY,
    job_id        BIGINT NOT NULL REFERENCES ingest_jobs(id) ON DELETE CASCADE,
    stage         TEXT NOT NULL CHECK (stage IN ('parse', 'embed', 'finalize')),
    file_key      TEXT,
    payload       JSONB NOT NULL DEFAULT '{}',
    status        TEXT NOT NULL DEFAULT 'pending' CHECK (status IN (
        'pending', 'processing', 'done', 'failed', 'dead'
    )),
    attempts      INT NOT NULL DEFAULT 0,
    max_attempts  INT NOT NULL DEFAULT 3,
    error_msg     TEXT,
    lease_until   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ingest_tasks_job ON ingest_tasks (job_id, stage, status);
CREATE INDEX IF NOT EXISTS idx_ingest_tasks_pending ON ingest_tasks (stage, status, created_at)
    WHERE status = 'pending';
