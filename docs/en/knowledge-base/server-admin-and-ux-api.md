# Admin and UX API

**Files:** `internal/admin/{handlers,routes,rbac}.go`, `internal/store/{analytics_store,audit_store,quota}.go`, `internal/domain/catalog.go`, `internal/locale/locale.go`, `internal/httpapi/*` (branding, onboarding, feedback, metrics)  
**Client:** [webapp-overview.md](./webapp-overview.md)

---

## Admin handlers (`internal/admin/handlers.go`)

### Authorization

HTTP Basic: `ADMIN_USER` / `ADMIN_PASSWORD`. Empty password → **503**.

### Routes `/admin` and `/api/admin`

| Method | Handler | Action |
|--------|---------|--------|
| GET | `handleAdminStatus` | `{ data_dir, domains }` |
| GET | `handleAdminListArticles` | files in `data/{tenant}/{domain}/` (legacy listing) |
| GET | `handleListKBDocuments` | documents from Postgres registry (`/kb/documents`) |
| POST | `handleAdminUpload` | dual-write: `data/` + blob + registry |
| DELETE | `handleAdminDeleteArticle` | delete file + soft-delete in registry |
| POST | `handleRebuildIndexRun` | create/activate index run (`/kb/index-runs`) |
| POST | `handleAdminReindex` | reindex via Python |
| POST | `handleIngest` | ingest job → Python pipeline |
| GET | `handleAdminFeedbackSummary` | aggregated thumbs up/down |
| GET | `handleAdminAuditLog` | admin audit trail (`?limit=&offset=&action=`) |
| GET | `handleAdminQuotas` | tenant quota limits + usage (`?tenant_id=`) |
| GET | `handleAdminAPIKeys` | API key labels + roles, admin user list (no secrets) |
| GET | `/admin/auth/*` | OIDC SSO login/callback/logout (public) |

### `GET /admin/articles`

Response: `articles[]` with `filename`, `size_bytes`, `modified`, `chunks` (from Python `/admin/index-stats`).

### Upload

- `domain_id`, optional `tenant_id`
- Formats: **`.txt`**, **`.pdf`**, **`.docx`**
- Regex: `^[a-zA-Z0-9._-]+\.(txt|pdf|docx)$`
- Max size: **10 MB**
- Writes: `{DATA_DIR}/{tenant_id}/{domain_id}/{filename}` **and** blob store **and** `kb_documents` / `kb_document_versions`
- Response includes `document_id`, `version_id` when registry is available

See [KB_SOURCE_OF_TRUTH.md](../KB_SOURCE_OF_TRUTH.md).

---

## `domains.go` — domain catalog

`loadDomainCatalog()` ← `DOMAINS_CONFIG_PATH` / `config/domains.json`

### `GET /domains`, `/api/domains`

Public, no Telegram auth. Display names use request locale (`names.ru` / `names.en`).

Response includes `locale` field.

---

## `onboarding.go`

`GET /onboarding?domain_id=default&locale=en` → `{ questions: [...], locale }`

Loaded from `config/locales/{locale}/onboarding.json`.

---

## `branding.go`

`GET /branding?locale=ru` → UI strings from `config/locales/{locale}/branding.json`

---

## `feedback.go`

`POST /feedback` — rating `1` / `-1` on an assistant message (Telegram or API key auth).

---

### `GET /admin/audit-log`

Query: `limit` (default 50, max 200), `offset`, optional `action` filter.

Response: `entries[]` with `occurred_at`, `action`, `actor`, `tenant_id`, `domain_id`, `resource`, `success`, `details`.

Actions: `admin_login`, `admin_login_failed`, `kb_upload`, `kb_delete`, `kb_reindex`, `kb_ingest`.

---

## Ingest chain (recommended)

Go `POST /admin/ingest` → creates `ingest_jobs` row → Python `POST /admin/ingest/run` → Redis workers (parse → embed → finalize).

`GET /admin/ingest/status?job_id=` — poll job + per-file tasks.

See [INGESTION.md](../INGESTION.md) and [data-pipeline.md](./data-pipeline.md).

### KB registry + index runs

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/admin/kb/documents?domain_id=` | Active documents from Postgres |
| POST | `/admin/kb/index-runs?domain_id=` | New index run; body `{"activate": true}` |

See [KB_SOURCE_OF_TRUTH.md](../KB_SOURCE_OF_TRUTH.md).

---

## Reindex chain (legacy / dev)

Go `POST /admin/reindex` → Python `POST /admin/reindex` + header `X-Admin-Secret`.

See [rag-vector_store.md](./rag-vector_store.md).

---

## Integrator API (Phase 2)

- `/api/v1/*` — versioned routes with OpenAPI spec
- `GET /metrics` — Prometheus text metrics (public or behind proxy)

See [server-auth-and-limits.md](./server-auth-and-limits.md).
