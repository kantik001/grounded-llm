# Примеры API

Базовый URL: `http://localhost:8080`  
OpenAPI: `GET /api/v1/openapi.json`

**Локально:** `TELEGRAM_AUTH_DISABLED=true` — без подписи Telegram.

Заголовки:

| Заголовок | Назначение |
|-----------|------------|
| `X-Telegram-Init-Data` | Авторизация Web App |
| `X-API-Key` | Интеграторы |
| `X-Tenant-ID` | Арендатор |
| `X-Locale` | `en` / `ru` |

Полный список (EN): [API_EXAMPLES.md](../en/API_EXAMPLES.md).

---

## Health / Ready

```bash
curl -sS http://localhost:8080/health
curl -sS http://localhost:8080/ready
```

---

## Домены и branding

```bash
curl -sS "http://localhost:8080/api/domains?locale=ru"
curl -sS "http://localhost:8080/api/branding?locale=ru"
curl -sS "http://localhost:8080/api/onboarding?domain_id=default&locale=ru"
```

---

## Сессия

```bash
curl -sS -X POST http://localhost:8080/api/session \
  -H "Content-Type: application/json; charset=utf-8" \
  -d '{"domain_id":"default"}'
```

---

## Сообщение в чат

```bash
SESSION_ID="<session_id>"
curl -sS -X POST http://localhost:8080/api/message \
  -H "Content-Type: application/json; charset=utf-8" \
  -d "{\"session_id\":\"${SESSION_ID}\",\"domain_id\":\"default\",\"text\":\"Сколько дней отпуска?\"}"
```

---

## Streaming (SSE)

```bash
curl -sS -N -X POST "http://localhost:8080/api/message?stream=1" \
  -H "Content-Type: application/json; charset=utf-8" \
  -H "Accept: text/event-stream" \
  -d "{\"session_id\":\"${SESSION_ID}\",\"domain_id\":\"default\",\"text\":\"Сколько дней отпуска?\"}"
```

---

## Admin upload

```bash
curl -sS -u admin:password \
  -F "domain_id=default" \
  -F "file=@./data/default/vacation_policy_en.txt" \
  http://localhost:8080/api/admin/upload
```

---

## Admin — квоты tenant

```bash
curl -sS -u admin:password "http://localhost:8080/api/admin/quotas?tenant_id=default"
```

---

## Опциональный SaaS signup

Требует `SAAS_SIGNUP_ENABLED=true`. См. [SAAS.md](./SAAS.md).

```bash
curl -sS http://localhost:8080/api/v1/plans

curl -sS -X POST http://localhost:8080/api/v1/signup \
  -H 'Content-Type: application/json' \
  -d '{"org_name":"Acme","email":"admin@acme.com","plan":"starter"}'
```

UI: `http://localhost/signup.html`

---

## Conformance (стандарт)

```bash
pip install -r conformance/requirements.txt
python -m conformance spec
python -m conformance check --url http://localhost:8080
```

---

## Smoke

```bash
bash scripts/smoke.sh http://localhost:8080
```

---

См. также: [DEPLOY.md](./DEPLOY.md), [knowledge-base/server-auth-and-limits.md](./knowledge-base/server-auth-and-limits.md).
