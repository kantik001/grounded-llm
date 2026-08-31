# PostgreSQL migrations

Versioned SQL applied **once** by the Go server on startup (`internal/app/main.go` → `store.RunAllMigrations`), recorded in `schema_migrations`.

Override directory: `MIGRATIONS_DIR` (Compose: `/migrations`).

| File | Purpose |
|------|---------|
| `001_init.sql` | users, chat_sessions, messages |
| `002_domain_id.sql` | session `domain_id` |
| `003_feedback_analytics.sql` | message_feedback, analytics_events |
| `005_message_citations.sql` | citations JSON on messages |
| `006_tenant_id.sql` | multi-tenant `tenant_id` |
| `007_audit_log.sql` | admin audit_log |
| `008_reindex_jobs.sql` | async reindex job queue |
| `009_pgvector.sql` | `CREATE EXTENSION vector` |
| `010_saas_tenants.sql` | saas_tenants, tenant_quotas, stripe_webhook_events |
| `011_admin_users_membership.sql` | admin_users, user_tenant_memberships |
| `012_ingest_jobs.sql` | async ingest jobs + tasks (parse/embed/index) |

There is no `004_*.sql` (numbering gap is historical; apply order is lexical sort of filenames).

## Rules

- New files: next number `NNN_short_name.sql`, idempotent where possible (`IF NOT EXISTS`).
- Never edit an already-applied migration on shared environments — add a new file.
- Deep dive: [docs/en/knowledge-base/migrations-overview.md](../docs/en/knowledge-base/migrations-overview.md).
