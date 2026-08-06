-- Sprint 4: admin users + Telegram tenant membership

CREATE TABLE IF NOT EXISTS admin_users (
    username         TEXT PRIMARY KEY,
    password_bcrypt  TEXT NOT NULL DEFAULT '',
    roles            TEXT[] NOT NULL DEFAULT '{}',
    tenant_id        TEXT NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_admin_users_tenant ON admin_users (tenant_id);

CREATE TABLE IF NOT EXISTS user_tenant_memberships (
    telegram_id  BIGINT NOT NULL,
    tenant_id    TEXT NOT NULL,
    role         TEXT NOT NULL DEFAULT 'member',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (telegram_id, tenant_id)
);

CREATE INDEX IF NOT EXISTS idx_user_tenant_memberships_tenant
    ON user_tenant_memberships (tenant_id);
