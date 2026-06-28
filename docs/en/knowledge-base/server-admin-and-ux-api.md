# Admin and UX API

**Files:** `admin.go`, `audit.go`, `audit_store.go`, `admin_feedback.go`, `domains.go`, `onboarding.go`, `branding.go`, `feedback.go`, `analytics_store.go`, `locale.go`  
**Client:** [webapp-overview.md](./webapp-overview.md)

---

## `admin.go` — knowledge base upload

### Authorization

HTTP Basic: `ADMIN_USER` / `ADMIN_PASSWORD`. Empty password → **503**.

### Routes `/admin` and `/api/admin`

| Method | Handler | Action |
|--------|---------|--------|
| GET | `handleAdminStatus` | `{ data_dir, domains }` |
| GET | `handleAdminListArticles` | files in `data/{tenant}/{domain}/` |
| POST | `handleAdminUpload` | save document |
| DELETE | `handleAdminDeleteArticle` | delete document (`?domain_id=&filename=&tenant_id=`) |
| POST | `handleAdminReindex` | reindex via Python |
| GET | `handleAdminFeedbackSummary` | aggregated thumbs up/down |
| GET | `handleAdminAuditLog` | admin audit trail (`?limit=&offset=&action=`) |
| GET | `handleAdminAPIKeys` | API key labels + roles, admin user list (no secrets) |

### `GET /admin/articles`

Response: `articles[]` with `filename`, `size_bytes`, `modified`, `chunks` (from Python `/admin/index-stats`).

### Upload

- `domain_id`, optional `tenant_id`
- Formats: **`.txt`**, **`.pdf`**, **`.docx`**
- Regex: `^[a-zA-Z0-9._-]+\.(txt|pdf|docx)$`
- Max size: **10 MB**
- Path: `{DATA_DIR}/{tenant_id}/{domain_id}/{filename}`

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

Actions: `admin_login`, `admin_login_failed`, `kb_upload`, `kb_delete`, `kb_reindex`.

---

## Reindex chain

Go `POST /admin/reindex` → Python `POST /admin/reindex` + header `X-Admin-Secret`.

See [rag-vector_store.md](./rag-vector_store.md).

---

## Integrator API (Phase 2)

- `/api/v1/*` — versioned routes with OpenAPI spec
- `GET /metrics` — Prometheus text metrics (public or behind proxy)

See [server-auth-and-limits.md](./server-auth-and-limits.md).
