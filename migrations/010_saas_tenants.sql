-- Sprint 3: SaaS identity in Postgres (tenants, quotas, Stripe webhook idempotency)

CREATE TABLE IF NOT EXISTS saas_tenants (
    tenant_id          TEXT PRIMARY KEY,
    org_name           TEXT NOT NULL DEFAULT '',
    email              TEXT NOT NULL DEFAULT '',
    plan               TEXT NOT NULL DEFAULT 'starter',
    stripe_customer_id TEXT NOT NULL DEFAULT '',
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_saas_tenants_email ON saas_tenants (lower(email));

CREATE TABLE IF NOT EXISTS tenant_quotas (
    tenant_id          TEXT PRIMARY KEY,
    messages_per_day   INT NOT NULL DEFAULT 200,
    storage_mb         INT NOT NULL DEFAULT 512,
    max_domains        INT NOT NULL DEFAULT 1,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS stripe_webhook_events (
    event_id     TEXT PRIMARY KEY,
    event_type   TEXT NOT NULL DEFAULT '',
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_stripe_webhook_events_processed
    ON stripe_webhook_events (processed_at DESC);
