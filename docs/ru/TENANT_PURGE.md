**Канон (EN):** [TENANT_PURGE.md](../en/TENANT_PURGE.md)

# Очистка данных tenant (RTBF)

**Статус:** реализовано — `DELETE /api/admin/tenants/:tenant_id`  
**Цель:** GDPR / right-to-be-forgotten для [TRUST_CENTER.md (EN)](../en/TRUST_CENTER.md)

Зачем: полный wipe арендатора (сессии, сообщения, файлы KB), а не только time-based retention.

---

## Endpoint

```http
DELETE /api/admin/tenants/{tenant_id}?confirm=true
Authorization: Basic (admin) или OIDC session с ролью `admin`
```

### Request

| Параметр | Required | Описание |
|----------|----------|----------|
| `tenant_id` | path | Tenant для purge (не `default` без доп. confirm) |
| `confirm` | query | Должен быть `true` |
| `purge_chroma` | query | Опц. `true` — удалить векторы tenant в Python (async job) |

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

### Ошибки

| Code | Условие |
|------|---------|
| `403` | Нужна роль `admin` |
| `400` | Нет `confirm=true` или невалидный tenant_id |
| `409` | Идёт reindex или ingest job для tenant |

---

## Что удаляется

| Store | Действие |
|-------|----------|
| Postgres `chat_sessions`, `messages`, `message_feedback` | DELETE WHERE tenant_id |
| Postgres `audit_log` | DELETE WHERE tenant_id (опц. анонимные агрегаты — config) |
| Postgres `kb_documents`, blobs | DELETE + удаление blob keys |
| Upload dir | Image tokens сессий tenant |
| Chroma | Filter-delete по metadata tenant (Python admin call) |

**Не удаляется:** `config/`, другие tenants, глобальные Prometheus counters.

---

## Audit

Перед удалением — запись:

```json
{
  "action": "tenant_purge",
  "tenant_id": "acme",
  "actor": "admin@example.com",
  "metadata": { "sessions": 42, "messages": 318 }
}
```

---

## Код (после strangler)

| Путь | Роль |
|------|------|
| `server/internal/admin/handlers_ops.go` | `handlePurgeTenant` + валидация |
| `server/internal/admin/routes.go` | `DELETE /tenants/:tenant_id` |
| `server/internal/store/tenant_purge_store.go` | `ChatStore.PurgeTenant` — SQL + файлы |
| `server/internal/audit/` | helpers audit log |

Checklist:

- [x] Handler + validation (RoleAdmin)
- [x] `ChatStore.PurgeTenant`
- [ ] Python `POST /admin/purge-tenant` — Chroma filter delete (опц. follow-up)
- [x] Tests: admin purge
- [ ] OpenAPI admin paths (follow-up)

---

## См. также

- Retention: `MESSAGE_RETENTION_DAYS` / `SESSION_RETENTION_DAYS` (по времени, не полный wipe)
- [PHASE_4.md (EN)](../en/PHASE_4.md)
- [BACKUP_RESTORE.md](./BACKUP_RESTORE.md)
