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

---

## Health

```bash
curl -sS http://localhost:8080/health
```

---

## Домены

```bash
curl -sS "http://localhost:8080/api/domains?locale=en"
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
  -d "{\"session_id\":\"${SESSION_ID}\",\"domain_id\":\"default\",\"text\":\"How many vacation days?\"}"
```

---

## Streaming (SSE)

```bash
curl -sS -N -X POST "http://localhost:8080/api/message?stream=1" \
  -H "Content-Type: application/json; charset=utf-8" \
  -H "Accept: text/event-stream" \
  -d "{\"session_id\":\"${SESSION_ID}\",\"domain_id\":\"default\",\"text\":\"How many vacation days?\"}"
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

## Smoke

```bash
bash scripts/smoke.sh http://localhost:8080
```

---

Полный список: [API_EXAMPLES.md](../en/API_EXAMPLES.md).
