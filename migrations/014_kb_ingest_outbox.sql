-- Outbox: registry upserts → async ingest enqueue (KB_AUTO_INGEST=1)

CREATE TABLE IF NOT EXISTS kb_ingest_outbox (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       TEXT NOT NULL,
    domain_id       TEXT NOT NULL,
    document_id     UUID NOT NULL,
    version_id      UUID NOT NULL,
    logical_key     TEXT NOT NULL,
    content_sha256  TEXT NOT NULL,
    source          TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'processing', 'done', 'failed')),
    ingest_job_id   BIGINT REFERENCES ingest_jobs(id) ON DELETE SET NULL,
    error_msg       TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_kb_ingest_outbox_pending
    ON kb_ingest_outbox (tenant_id, domain_id, created_at)
    WHERE status = 'pending';
