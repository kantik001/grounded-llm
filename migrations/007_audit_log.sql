-- Phase B: minimal admin audit log (KB mutations + auth events)

CREATE TABLE IF NOT EXISTS audit_log (
    id          BIGSERIAL PRIMARY KEY,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    action      TEXT NOT NULL,
    actor       TEXT,
    tenant_id   TEXT,
    domain_id   TEXT,
    resource    TEXT,
    client_ip   TEXT,
    request_id  TEXT,
    success     BOOLEAN NOT NULL DEFAULT TRUE,
    details     JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_audit_log_occurred ON audit_log (occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_action_time ON audit_log (action, occurred_at DESC);
