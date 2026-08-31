-- Enterprise KB source of truth: document registry, versions, ACL, index runs

CREATE TABLE IF NOT EXISTS kb_documents (
    id              UUID PRIMARY KEY,
    tenant_id       TEXT NOT NULL,
    domain_id       TEXT NOT NULL,
    logical_key     TEXT NOT NULL,
    title           TEXT NOT NULL DEFAULT '',
    mime_type       TEXT NOT NULL DEFAULT 'application/octet-stream',
    status          TEXT NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'deleted', 'archived')),
    current_version INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, domain_id, logical_key)
);

CREATE INDEX IF NOT EXISTS idx_kb_documents_scope
    ON kb_documents (tenant_id, domain_id, status);

CREATE TABLE IF NOT EXISTS kb_document_versions (
    id              UUID PRIMARY KEY,
    document_id     UUID NOT NULL REFERENCES kb_documents(id) ON DELETE CASCADE,
    version         INT NOT NULL,
    storage_key     TEXT NOT NULL,
    content_sha256  TEXT NOT NULL,
    size_bytes      BIGINT NOT NULL DEFAULT 0,
    source          TEXT NOT NULL DEFAULT 'upload',
    source_ref      JSONB NOT NULL DEFAULT '{}',
    created_by      TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (document_id, version)
);

CREATE INDEX IF NOT EXISTS idx_kb_document_versions_doc
    ON kb_document_versions (document_id, version DESC);
CREATE INDEX IF NOT EXISTS idx_kb_document_versions_sha
    ON kb_document_versions (content_sha256);

CREATE TABLE IF NOT EXISTS kb_document_acl (
    document_id     UUID NOT NULL REFERENCES kb_documents(id) ON DELETE CASCADE,
    principal_type  TEXT NOT NULL CHECK (principal_type IN ('tenant', 'role', 'user', 'group')),
    principal_id    TEXT NOT NULL,
    permission      TEXT NOT NULL DEFAULT 'read' CHECK (permission IN ('read', 'admin')),
    PRIMARY KEY (document_id, principal_type, principal_id)
);

CREATE INDEX IF NOT EXISTS idx_kb_document_acl_principal
    ON kb_document_acl (principal_type, principal_id);

-- Disposable index generations (blue/green rebuilds)
CREATE TABLE IF NOT EXISTS index_runs (
    id              UUID PRIMARY KEY,
    tenant_id       TEXT,
    domain_id       TEXT,
    backend         TEXT NOT NULL DEFAULT 'chroma',
    embedding_model TEXT NOT NULL DEFAULT '',
    chunk_schema    JSONB NOT NULL DEFAULT '{}',
    status          TEXT NOT NULL DEFAULT 'building'
        CHECK (status IN ('building', 'active', 'retired', 'failed')),
    error_msg       TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    activated_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_index_runs_scope
    ON index_runs (tenant_id, domain_id, status, created_at DESC);

-- Per-document indexing state within an index run
CREATE TABLE IF NOT EXISTS index_document_state (
    index_run_id      UUID NOT NULL REFERENCES index_runs(id) ON DELETE CASCADE,
    document_id       UUID NOT NULL REFERENCES kb_documents(id) ON DELETE CASCADE,
    indexed_version   INT NOT NULL,
    content_sha256    TEXT NOT NULL,
    chunk_count       INT NOT NULL DEFAULT 0,
    indexed_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (index_run_id, document_id)
);

CREATE INDEX IF NOT EXISTS idx_index_document_state_doc
    ON index_document_state (document_id);

-- Active index run pointer per tenant+domain (NULL scope = global default)
CREATE TABLE IF NOT EXISTS index_run_active (
    tenant_id       TEXT NOT NULL DEFAULT '',
    domain_id       TEXT NOT NULL DEFAULT '',
    index_run_id    UUID NOT NULL REFERENCES index_runs(id) ON DELETE RESTRICT,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, domain_id)
);
