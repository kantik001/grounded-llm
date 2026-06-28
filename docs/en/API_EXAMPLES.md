# API Examples

Base URL: `http://localhost:8080` (or your deployment).  
OpenAPI spec: `GET /api/v1/openapi.json`

**Local dev:** set `TELEGRAM_AUTH_DISABLED=true` to skip Telegram signature on protected routes.

Common headers:

| Header | Purpose |
|--------|---------|
| `X-Telegram-Init-Data` | Telegram Web App auth |
| `X-API-Key` | Integrator auth (alternative) |
| `X-Tenant-ID` | Multi-tenant isolation |
| `X-Locale` | `en` or `ru` |
| `Content-Type` | `application/json; charset=utf-8` |

---

## Health

```bash
curl -sS http://localhost:8080/health
```

---

## Domains catalog

```bash
curl -sS "http://localhost:8080/api/domains?locale=en"
```

---

## Branding (UI strings)

```bash
curl -sS "http://localhost:8080/api/branding?locale=en"
```

---

## Onboarding sample questions

```bash
curl -sS "http://localhost:8080/api/onboarding?domain_id=default&locale=en"
```

---

## Create session

```bash
curl -sS -X POST http://localhost:8080/api/session \
  -H "Content-Type: application/json; charset=utf-8" \
  -d '{"domain_id":"default"}'
```

Response: `{"success":true,"session_id":"...","domain_id":"default"}`

---

## Chat history

```bash
SESSION_ID="<from create session>"
curl -sS "http://localhost:8080/api/history?session_id=${SESSION_ID}"
```

---

## Send message (JSON)

```bash
curl -sS -X POST http://localhost:8080/api/message \
  -H "Content-Type: application/json; charset=utf-8" \
  -d "{\"session_id\":\"${SESSION_ID}\",\"domain_id\":\"default\",\"text\":\"How many vacation days?\"}"
```

---

## Send message (SSE stream)

```bash
curl -sS -N -X POST "http://localhost:8080/api/message?stream=1" \
  -H "Content-Type: application/json; charset=utf-8" \
  -H "Accept: text/event-stream" \
  -d "{\"session_id\":\"${SESSION_ID}\",\"domain_id\":\"default\",\"text\":\"How many vacation days?\"}"
```

---

## Message feedback

```bash
MESSAGE_ID=123
curl -sS -X POST http://localhost:8080/api/feedback \
  -H "Content-Type: application/json; charset=utf-8" \
  -d "{\"message_id\":${MESSAGE_ID},\"rating\":1}"
```

---

## API key + tenant (integrators)

```bash
curl -sS -X POST http://localhost:8080/api/v1/session \
  -H "X-API-Key: your-key" \
  -H "X-Tenant-ID: acme" \
  -H "Content-Type: application/json; charset=utf-8" \
  -d '{"domain_id":"default"}'
```

---

## Admin — status (Basic Auth)

```bash
curl -sS -u admin:your-password http://localhost:8080/api/admin/status
```

---

## Admin — upload document

```bash
curl -sS -u admin:your-password \
  -F "domain_id=default" \
  -F "file=@./data/default/vacation_policy_en.txt" \
  http://localhost:8080/api/admin/upload
```

---

## Admin — reindex RAG

```bash
curl -sS -u admin:your-password -X POST http://localhost:8080/api/admin/reindex
```

---

## Admin — audit log

```bash
curl -sS -u admin:your-password "http://localhost:8080/api/admin/audit-log?limit=20"
```

---

## Admin — tenant quotas

```bash
curl -sS -u admin:your-password "http://localhost:8080/api/admin/quotas?tenant_id=default"
```

---

## Metrics (Prometheus)

```bash
curl -sS http://localhost:8080/metrics
```

Protect this endpoint in production.

---

## Smoke test script

```bash
bash scripts/smoke.sh http://localhost:8080
```

---

See also: [PILOT_PLAYBOOK.md](./PILOT_PLAYBOOK.md), [server-auth-and-limits.md](./knowledge-base/server-auth-and-limits.md).
