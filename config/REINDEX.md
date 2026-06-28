# Async RAG reindex

Reindex rebuilds Chroma from all files under `data/`. The admin API runs reindex **asynchronously** so the UI and HTTP client are not blocked for minutes.

## API

| Method | Path | Role | Description |
|--------|------|------|-------------|
| POST | `/api/admin/reindex` | `kb_editor` | Queue a reindex job; returns immediately |
| GET | `/api/admin/reindex/status?job_id=` | `kb_editor` | Poll job status (`job_id` optional — latest active or most recent) |

### POST response (202 Accepted)

```json
{
  "success": true,
  "job_id": 42,
  "status": "pending",
  "status_label": "queued",
  "already_running": false,
  "message": "RAG reindex queued"
}
```

If a job is already pending/running, the same shape is returned with `already_running: true` and the existing `job_id`.

### GET status response

```json
{
  "success": true,
  "done": false,
  "status_label": "running",
  "job": {
    "id": 42,
    "status": "running",
    "created_at": "2026-06-29T12:00:00Z",
    "started_at": "2026-06-29T12:00:01Z"
  }
}
```

Job statuses: `pending` → `running` → `succeeded` | `failed`.

## Behaviour

- Only **one** active reindex at a time (global Chroma rebuild).
- Go worker calls Python `POST /admin/reindex` (unchanged); Python still runs synchronously inside the worker goroutine.
- Audit log entry `kb_reindex` is written when the job **finishes** (success or failure), with `job_id` in details.
- Jobs are stored in Postgres (`reindex_jobs`, migration `008_reindex_jobs.sql`).

## UI

Admin panel polls `/reindex/status` every 2 seconds after starting a job and resumes polling if you reload the page while a job is active.

## CLI / scripts

`python scripts/reindex_rag.py` and CI eval gate remain synchronous — they do not use the job queue.
