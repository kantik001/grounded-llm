-- Multi-tenant workspace isolation (Phase 2)

ALTER TABLE chat_sessions
    ADD COLUMN IF NOT EXISTS tenant_id TEXT NOT NULL DEFAULT 'default';

CREATE INDEX IF NOT EXISTS idx_chat_sessions_tenant_id ON chat_sessions (tenant_id);
