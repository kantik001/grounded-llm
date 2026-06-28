# Per-tenant quotas

Optional caps per tenant for **messages/day**, **KB storage**, and **number of domains** with documents.

## Config

```bash
TENANT_QUOTAS_FILE=config/tenant_quotas.json
```

Example: [tenant_quotas.json.example](./tenant_quotas.json.example)

```json
[
  {
    "tenant_id": "default",
    "messages_per_day": 10000,
    "storage_mb": 512,
    "max_domains": 20
  },
  {
    "tenant_id": "acme",
    "messages_per_day": 500,
    "storage_mb": 100,
    "max_domains": 3
  }
]
```

**`0` or omitted field** = unlimited for that metric.  
**No file** = quotas disabled (backward compatible).

Reload: `SIGHUP` or `CONFIG_RELOAD_INTERVAL_SEC`.

## Enforcement

| Quota | Checked on | HTTP |
|-------|------------|------|
| `messages_per_day` | `POST /message` (user messages, UTC day) | 429 |
| `storage_mb` | `POST /admin/upload` | 413 |
| `max_domains` | `POST /admin/upload` to a new domain dir | 400 |

## Usage API

```bash
curl -u admin:password "http://localhost:8080/api/admin/quotas?tenant_id=default"
```

Response includes `limits`, `usage` (`messages_today`, `storage_mb`, `domains`), and `enforced`.

Requires `kb_editor` or `admin` role.

## Notes

- Message counts use Postgres (`messages` + `chat_sessions.tenant_id`).
- Storage counts `.txt`/`.pdf`/`.docx` under `data/{tenant_id}/`.
- Domain count = distinct KB directories with at least one document.
