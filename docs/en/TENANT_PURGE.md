# Tenant data purge (RTBF)

**Status:** implemented — `DELETE /api/admin/tenants/:tenant_id`  
**Goal:** GDPR / right-to-be-forgotten story for [TRUST_CENTER.md](./TRUST_CENTER.md)

---

## Endpoint

```http
DELETE /api/admin/tenants/{tenant_id}?confirm=true
Authorization: Basic (admin) or OIDC session with role `admin`
```

### Request

| Parameter | Required | Description |
|-----------|----------|-------------|
| `tenant_id` | path | Tenant to purge (not `default` without extra confirm) |
| `confirm` | query | Must be `true` |
| `purge_chroma` | query | Optional `true` — remove tenant vectors in Python (async job) |

### Response `200`

```json
{
  "success": true,
  "tenant_id": "acme",
  "deleted": {
    "sessions": 42,
    "messages": 318,
    "feedback_rows": 12,
    "audit_rows": 0,
    "data_files": 5,
    "upload_tokens": 2
  }
}
```

### Errors

| Code | Condition |
|------|-----------|
| `403` | Insufficient role (requires `admin`) |
| `400` | Missing `confirm=true` or invalid tenant_id |
| `409` | Reindex job running for tenant |

---

## Data removed

| Store | Action |
|-------|--------|
| Postgres `chat_sessions`, `messages`, `message_feedback` | DELETE WHERE tenant_id |
| Postgres `audit_log` | DELETE WHERE tenant_id (optional retain anonymized aggregate — config) |
| Filesystem `data/{tenant_id}/` | Recursive delete |
| Upload dir | Delete image tokens linked to purged sessions |
| Chroma | Filter-delete by tenant metadata (Python admin call) |

**Not deleted:** platform config (`config/`), other tenants, global metrics counters.

---

## Audit

Before deletion, append audit row:

```json
{
  "action": "tenant_purge",
  "tenant_id": "acme",
  "actor": "admin@example.com",
  "metadata": { "sessions": 42, "messages": 318 }
}
```

---

## Implementation checklist (Phase 4)

- [x] `server/admin_tenant_purge.go` — handler + validation
- [x] `ChatStore.PurgeTenant` — SQL + file cleanup
- [ ] Python `POST /admin/purge-tenant` — Chroma filter delete (optional follow-up)
- [x] RBAC: `RoleAdmin` only
- [x] Tests: `admin_tenant_purge_test.go`
- [ ] OpenAPI admin paths extension (follow-up)
- [x] Trust Center updated

---

## Related

- Retention worker: `MESSAGE_RETENTION_DAYS` / `SESSION_RETENTION_DAYS` (time-based, not full tenant delete)
- [PHASE_4.md](./PHASE_4.md)
