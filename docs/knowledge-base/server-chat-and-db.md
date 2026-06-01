# Разбор: чат и БД / Chat & database

**Файлы / Files:** `message_handlers.go`, `session_handlers.go`, `chat_session.go`, `postgres_store.go`  
**Схема / Schema:** [migrations-overview.md](./migrations-overview.md)

---

## `POST /message`

Auth (Telegram) + rate limit.

JSON:

```json
{ "session_id": "...", "text": "...", "domain_id": "default" }
```

**Только текст** — RAG через `answerWithRAG`.  
**Multipart с image** — отклоняется с сообщением «vision module not in core» (подключите domain pack).

---

## Сессии / Sessions

- `POST /session` — новая `chat_sessions` + `session_id`
- `GET /history?session_id=` — сообщения
- Колонка `domain_id` в `chat_sessions` (миграция `004_domain_id.sql`)

---

## Postgres tables

| Таблица | Назначение |
|---------|------------|
| `users` | Telegram users |
| `chat_sessions` | session + `domain_id` |
| `messages` | user / assistant |
| `message_feedback` | 👍/👎 |
| `analytics_events` | product events |

---

## Ответ API

```json
{ "success": true, "session_id": "...", "domain_id": "...", "messages": [...] }
```

---

## Связанные docs

| Тема | Файл |
|------|------|
| RAG | [server-rag_chat.md](./server-rag_chat.md) |
| Auth | [server-auth-and-limits.md](./server-auth-and-limits.md) |
