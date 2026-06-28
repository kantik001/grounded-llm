# RBAC (Role-Based Access Control)

Phase B minimal RBAC: file-based roles for **admin users** (HTTP Basic Auth) and **API keys**.

## Roles

| Role | Admin API | Chat API (`X-API-Key`) |
|------|-----------|-------------------------|
| `chat_only` | — | session, message, history, feedback |
| `kb_editor` | list/upload/delete KB, reindex | — |
| `admin` | all admin routes (feedback, audit log) | — |
| `api_manager` | `GET /admin/api-keys` (labels + roles) | — |

`admin` is a **superuser** role for admin routes.

Telegram Web App users keep implicit `chat_only` (unchanged).

## Admin users

**Legacy (single user):** `ADMIN_USER` + `ADMIN_PASSWORD` in `.env` → role `admin`.

**Multi-user:** `ADMIN_USERS_FILE=config/admin_users.json`

```json
[
  {
    "username": "admin",
    "password_bcrypt": "$2a$10$...",
    "roles": ["admin"]
  },
  {
    "username": "editor",
    "password": "changeme",
    "roles": ["kb_editor"]
  }
]
```

Use `password_bcrypt` in production (`bcrypt` cost 10). Plain `password` is for local dev only.

Example file: [admin_users.json.example](./admin_users.json.example) (admin password: `password`).

## API keys

Extend [api_keys.json.example](./api_keys.json.example):

```json
[
  { "key": "secret", "label": "bot", "roles": ["chat_only"] }
]
```

Keys without `roles` default to `chat_only` (backward compatible).

`API_KEYS=key:label` env format still works (always `chat_only`).

## Reload

Send `SIGHUP` to the Go server or set `CONFIG_RELOAD_INTERVAL_SEC` to reload domains, locales, API keys, and admin users.

## Verify

```bash
# kb_editor — upload OK, audit log forbidden
curl -u editor:changeme -X POST -F "domain_id=default" -F "file=@doc.txt" http://localhost:8080/api/admin/upload
curl -u editor:changeme http://localhost:8080/api/admin/audit-log   # 403

# api_manager — keys list OK, upload forbidden
curl -u api_manager:changeme http://localhost:8080/api/admin/api-keys
```

See also: [docs/en/knowledge-base/server-auth-and-limits.md](../docs/en/knowledge-base/server-auth-and-limits.md)
