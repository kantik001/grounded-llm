# Чат и база данных

**Файлы:** `message_handlers.go`, `session_handlers.go`, `sse.go`, `postgres_store.go`  
**Схема:** [migrations-overview.md](./migrations-overview.md)

---

## `POST /message`

Авторизация (Telegram или `X-API-Key`) + rate limit.

JSON:

```json
{
  "session_id": "...",
  "text": "...",
  "domain_id": "default",
  "tenant_id": "optional"
}
```

**Только текст** — RAG через `answerWithRAG`; у ответа ассистента в UI — `citations[]` (фрагменты KB).  
**Потоковый ответ:** `POST /message?stream=1` — SSE с токенами (Web App с fallback на JSON).  
**Multipart с image** — отклоняется («vision module not in core»; нужен domain pack).

Язык промптов: `X-Locale`, `Accept-Language` или `?locale=`.

---

## Сессии

- `POST /session` — новая строка `chat_sessions` + `session_id`
- `GET /history?session_id=` — сообщения
- Колонки: `domain_id` (миграция `002`), `tenant_id` (миграция `006`)

---

## Таблицы PostgreSQL

| Таблица | Назначение |
|---------|------------|
| `users` | Пользователи Telegram |
| `chat_sessions` | Сессия + `domain_id` + `tenant_id` |
| `messages` | user / assistant (+ `citations JSONB`) |
| `message_feedback` | 👍 / 👎 |
| `analytics_events` | События продукта |

---

## Ответ API

```json
{
  "success": true,
  "session_id": "...",
  "domain_id": "...",
  "tenant_id": "...",
  "messages": [...]
}
```

---

## Связанные статьи

| Тема | Файл |
|------|------|
| RAG | [server-rag_chat.md](./server-rag_chat.md) |
| Auth | [server-auth-and-limits.md](./server-auth-and-limits.md) |
