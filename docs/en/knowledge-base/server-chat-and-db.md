# Chat and database

**Files:** `message_handlers.go`, `session_handlers.go`, `chat_session.go`, `postgres_store.go`, `sse.go`  
**Schema:** [migrations-overview.md](./migrations-overview.md)

---

## `POST /message`

Auth (Telegram initData or `X-API-Key`) + rate limit.

JSON:

```json
{ "session_id": "...", "text": "...", "domain_id": "default", "tenant_id": "optional" }
```

**Text only** — RAG via `answerWithRAG`; assistant replies include `citations[]` (KB fragments).  
**Streaming:** `POST /message?stream=1` — SSE token stream (Web App uses this with JSON fallback).  
**Multipart with image** — rejected with “vision module not in core” (use a domain pack).

Locale for prompts: `X-Locale`, `Accept-Language`, or `?locale=`.

---

## Sessions

- `POST /session` — new `chat_sessions` row + `session_id`
- `GET /history?session_id=` — messages
- Columns: `domain_id` (migration `002`), `tenant_id` (migration `006`)

---

## Postgres tables

| Table | Purpose |
|-------|---------|
| `users` | Telegram users |
| `chat_sessions` | session + `domain_id` + `tenant_id` |
| `messages` | user / assistant (+ `citations JSONB`) |
| `message_feedback` | thumbs up/down |
| `analytics_events` | product events |

---

## API response

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

## Related docs

| Topic | File |
|-------|------|
| RAG | [server-rag_chat.md](./server-rag_chat.md) |
| Auth | [server-auth-and-limits.md](./server-auth-and-limits.md) |
