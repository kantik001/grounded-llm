# Ingestion pipeline

Async KB ingestion: **parse → embed → index**, with job state in Postgres and work items in Redis.

See also: [KB_SOURCE_OF_TRUTH.md](./KB_SOURCE_OF_TRUTH.md) · [CONNECTORS.md](./CONNECTORS.md) · admin reindex (`POST /admin/reindex`) as full-sync fallback.

## Architecture

**Document source (production):** Postgres registry (`kb_documents`) + blob store (`KB_BLOB_DIR` or S3).

```
Upload / connector
    → kb_documents + kb_document_versions (Postgres)
    → blobs (local or S3/MinIO)
        ↓
POST /admin/ingest (Go) → ingest_jobs (Postgres)
        ↓
POST /admin/ingest/run (Python) → Redis queues
        ↓
discover_targets()  →  Postgres registry
parse worker        →  blob or local path → staging/*.chunks.jsonl
embed worker        →  Chroma (+ index_document_state)
finalize            →  BM25 rebuild + index run pointer
```

Retrieve path (`/rag/context`) is unchanged and does not block on ingestion. Set `KB_ACL_ENFORCE=1` to filter hits by `kb_document_acl`.

## Rollout (weeks)

| Week | Deliverable |
|------|-------------|
| 1 | `ingest_jobs` table, Go enqueue API, sync in-process worker (`sync=true`) |
| 2 | Redis queues, parse stage, chunk staging on disk |
| 3 | Embed worker, per-file Chroma upsert (`upsert_kb_file`) |
| 4 | Status API, retries, DLQ, Prometheus-style metrics |
| 5+ | KB registry + blobs, index runs, connector registry sync |

## API

### Go (admin, KB editor role)

```http
POST /admin/ingest?domain_id=hr
Content-Type: application/json

{"files": ["policy.pdf"], "mode": "incremental", "sync": false}
```

```http
GET /admin/ingest/status?job_id=42
GET /admin/ingest/status?domain_id=hr
GET /admin/kb/documents?domain_id=hr
POST /admin/kb/index-runs?domain_id=hr
{"activate": true, "backend": "chroma", "embedding_model": "intfloat/multilingual-e5-small"}
```

Empty `files` in ingest = all active documents in the domain (from Postgres registry).

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
| `DATABASE_URL` | — | Job/task state + KB registry (required) |
| `REDIS_URL` | — | Queues (required for async) |
| `KB_BLOB_BACKEND` | `local` | Blob store: `local` or `s3` |
| `KB_BLOB_DIR` | `{DATA_DIR}/blobs` | Local blob root |
| `KB_S3_*` | — | S3/MinIO when `KB_BLOB_BACKEND=s3` |
| `KB_ACL_ENFORCE` | `0` | Filter retrieval by document ACL |
| `KB_REGISTRY_SYNC` | `1` | Auto-register blobs after connector sync |

See [KB_SOURCE_OF_TRUTH.md](./KB_SOURCE_OF_TRUTH.md) for full SoT env list.

## Local dev

```bash
# Backfill legacy data/ into registry (once)
python scripts/backfill_kb_registry.py

# Sync (no worker container)
curl -X POST "http://localhost:8080/admin/ingest?domain_id=default" \
  -H "X-Admin-Secret: $ADMIN_SECRET" \
  -d '{"sync": true}'

# Async with worker
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
