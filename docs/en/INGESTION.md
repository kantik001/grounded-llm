# Ingestion pipeline

Async KB ingestion: **parse → embed → index**, with job state in Postgres and work items in Redis.

See also: [CONNECTORS.md](./CONNECTORS.md) · admin reindex (`POST /admin/reindex`) as full-sync fallback.

## Architecture

```
Upload / connector sync → data/{tenant}/{domain}/
        ↓
POST /admin/ingest (Go) → ingest_jobs (Postgres)
        ↓
POST /admin/ingest/run (Python) → Redis queues
        ↓
parse worker  → staging/*.chunks.jsonl
embed worker  → Chroma (+ manifest)
finalize      → BM25 rebuild
```

Retrieve path (`/rag/context`) is unchanged and does not block on ingestion.

## Rollout (weeks)

| Week | Deliverable |
|------|-------------|
| 1 | `ingest_jobs` table, Go enqueue API, sync in-process worker (`sync=true`) |
| 2 | Redis queues, parse stage, chunk staging on disk |
| 3 | Embed worker, per-file Chroma upsert (`upsert_kb_file`) |
| 4 | Status API, retries, DLQ, Prometheus-style metrics |
| 5+ | Scale workers, connector cron → auto enqueue |

## API

### Go (admin, KB editor role)

```http
POST /admin/ingest?domain_id=hr
Content-Type: application/json
X-Admin-Secret: ...

{"files": ["policy.pdf"], "mode": "incremental", "sync": false}
```

```http
GET /admin/ingest/status?job_id=42
GET /admin/ingest/status?domain_id=hr
```

### Python (internal)

```http
POST /admin/ingest/run   {"job_id": 42, "sync": false}
GET  /admin/ingest/status?job_id=42
GET  /admin/ingest/metrics
```

## Environment

| Variable | Default | Purpose |
|----------|---------|---------|
| `INGEST_ASYNC` | `1` | Use Redis queues; `0` = drain in API process |
| `INGEST_USE_REDIS` | `1` | Worker transport |
| `INGEST_STAGING_DIR` | `ingest_staging` | Parsed chunk JSONL |
| `INGEST_WORKER_STAGE` | `all` | `parse`, `embed`, `finalize`, or `all` |
| `DATABASE_URL` | — | Job/task state (required) |
| `REDIS_URL` | — | Queues (required for async) |

## Local dev

```bash
# Week 1 — sync (no worker container)
curl -X POST "http://localhost:8080/admin/ingest?domain_id=default" \
  -H "X-Admin-Secret: $ADMIN_SECRET" \
  -d '{"sync": true}'

# Week 2+ — async with worker
docker compose up -d ingest-worker
curl -X POST "http://localhost:8080/admin/ingest?domain_id=default" \
  -H "X-Admin-Secret: $ADMIN_SECRET" \
  -d '{"sync": false}'
```

## Reliability

- Tasks: `pending → processing → done | failed | dead`
- Retries: up to 3 attempts, then `ingest:dlq`
- Stale lease requeue after worker crash
- Job terminal states: `succeeded`, `failed`, `partial`

Legacy `POST /admin/reindex` remains for full/incremental sync without the job pipeline.
