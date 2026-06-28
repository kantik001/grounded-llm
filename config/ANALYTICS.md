# Analytics dashboard

Admin-only metrics for product owners: usage, RAG quality, and KB gaps.

## API

`GET /api/admin/analytics` (requires `admin` role)

| Query | Default | Description |
|-------|---------|-------------|
| `tenant_id` | *(empty)* | Filter by tenant; empty = all tenants |
| `days` | `7` | Window 1–90 days |

### Response highlights

- **questions_total** / **questions_today** — user messages in chat
- **questions_per_day** — daily breakdown
- **rag.verify_pass_rate** — % of answers that passed source verification (excludes soft-fail / no-context cases)
- **rag.soft_fail** — questions with no usable KB context
- **feedback** — thumbs up/down totals
- **top_domains** — busiest domains by question count
- **kb_gaps** — recent soft-fail and verify-fail questions (80-char preview, no full PII)

## Data sources

| Metric | Source |
|--------|--------|
| Questions | `messages` + `chat_sessions` |
| Verify rate / KB gaps | `analytics_events` (`event_type = rag_answer`) |
| Feedback | `message_feedback` |

RAG outcomes are recorded automatically on each chat reply (REST and SSE). Events are written only for real RAG attempts (context found or soft-fail), not for LLM/config errors.

## UI

Open **Admin → Analytics** (`webapp/admin.html`) after signing in with an account that has the `admin` role.

## Privacy

Question previews in KB gaps are truncated to 80 characters. Do not store secrets in KB documents if previews could expose them.
