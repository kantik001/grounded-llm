# Enterprise KB source of truth

Postgres holds **document metadata + ACL**; object storage holds **versioned blobs**; Chroma/BM25 are **disposable indexes**.

## Architecture

```text
Upload / connector
    → kb_documents + kb_document_versions (Postgres)
    → blobs in local KB_BLOB_DIR or S3 (MinIO)
    → ingest pipeline → Chroma + BM25 (rebuild anytime)
```

## Environment

| Variable | Default | Purpose |
|----------|---------|---------|
| `KB_BLOB_BACKEND` | `local` | `local` or `s3` |
| `KB_BLOB_DIR` | `{DATA_DIR}/blobs` | Local blob root |
| `KB_S3_ENDPOINT` | — | MinIO/S3 endpoint (e.g. `minio:9000`) |
| `KB_S3_BUCKET` | `grounded-kb` | Bucket name |
| `KB_S3_PREFIX` | — | Optional key prefix |
| `KB_S3_REGION` | `us-east-1` | S3 region |
| `KB_S3_USE_SSL` | `false` | HTTPS to endpoint |
| `KB_S3_ACCESS_KEY` / `KB_S3_SECRET_KEY` | — | Credentials |
| `KB_ACL_ENFORCE` | `0` | Filter retrieval by `kb_document_acl` |
| `KB_REGISTRY_SYNC` | `1` | Register blobs after connector sync (Google Drive) |
| `KB_AUTO_INGEST` | `0` | Write `kb_ingest_outbox` on upsert; flush → ingest job |

Optional MinIO in Compose:

```bash
docker compose --profile minio up -d
# KB_BLOB_BACKEND=s3 KB_S3_ENDPOINT=minio:9000 KB_S3_ACCESS_KEY=minioadmin ...
```

## Admin API

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/admin/kb/documents?domain_id=` | List active documents from Postgres |
| POST | `/admin/kb/index-runs?domain_id=` | Create index run; `"activate": true` to flip blue/green pointer |
| POST | `/admin/upload` | Writes blob + registry; with `KB_AUTO_INGEST=1` also queues ingest (`ingest_job_id`) |
| DELETE | `/admin/articles?filename=` | Soft-deletes document in registry (`status=deleted`) |

## Migration

1. Apply migration `013_kb_documents.sql` (automatic on server start).
2. Backfill existing files:

```bash
python scripts/backfill_kb_registry.py
python scripts/backfill_kb_registry.py --tenant default --domain default --dry-run
```

3. Run ingest to rebuild indexes from registry:

```bash
curl -u admin:pass -X POST "http://localhost:8080/api/admin/ingest?domain_id=default" \
  -H "Content-Type: application/json" -d '{"sync": true}'
```

## Index runs (blue/green)

Each tenant+domain has an **active index run** pointer in `index_run_active`. Vector collections (Chroma / Qdrant / pgvector) and BM25 pickles are **namespaced by run id**; ingest writes `index_document_state` per document.

### Correct workflow (zero-downtime rebuild)

```bash
# 1. Create a building run (do NOT activate yet)
curl -u admin:pass -X POST "http://localhost:8080/api/admin/kb/index-runs?domain_id=default" \
  -H "Content-Type: application/json" \
  -d '{"backend": "chroma", "embedding_model": "intfloat/multilingual-e5-small"}'

# 2. Full ingest INTO the building run; flip pointer when done
curl -u admin:pass -X POST "http://localhost:8080/api/admin/ingest?domain_id=default" \
  -H "Content-Type: application/json" \
  -d '{"mode": "full", "index_run_id": "<run-uuid>", "activate_on_complete": true}'

# 3. (Optional) GC old retired runs
python scripts/gc_index_runs.py --tenant default --domain default --keep-last 1
```

| Field | Purpose |
|-------|---------|
| `index_run_id` | Target building run for ingest writes (separate collection) |
| `activate_on_complete` | Atomically flip `index_run_active` + rebuild BM25 after ingest |

Incremental ingest (no `index_run_id`) writes to the **active** run and rebuilds sparse index for that scope.

Chunk IDs stay positional (`{tenant}/{domain}/{filename}/{seq}`); version metadata (`document_version`, `content_sha256`) is stored on chunk metadata, not in `chunk_id`.

Upload and first ingest also call `EnsureActiveIndexRun` for the scope when no run exists yet.

## Connectors

With `KB_REGISTRY_SYNC=1` (default), **Google Drive** registers synced files via `connectors/registry_sync.py` after staging download.

For other connectors, `scripts/sync_connector.py` registers via `register_synced_tree(...)` automatically. With `KB_AUTO_INGEST=1`, the CLI flushes `kb_ingest_outbox` via `POST /admin/ingest`. One-time migration from old `data/` trees: `scripts/backfill_kb_registry.py` (keep `KB_AUTO_INGEST=0`).

## Code layout

| Path | Role |
|------|------|
| `migrations/013_kb_documents.sql` | Schema |
| `migrations/014_kb_ingest_outbox.sql` | Ingest outbox (auto-enqueue) |
| `server/internal/store/kb_documents.go` | Go registry + ACL |
| `server/internal/store/kb_outbox.go` | Go outbox flush |
| `rag/kb/outbox.py` | Python outbox + HTTP flush |
| `server/internal/kb/blobstore/` | Go blob backend |
| `rag/kb/documents.py` | Python registry + discover |
| `rag/kb/index_runs.py` | Index run helpers |
| `rag/kb/index_collections.py` | Per-run collection / persist path naming |
| `scripts/gc_index_runs.py` | GC retired Chroma + sparse run directories |
| `rag/storage/blob_store.py` | Python blob backend |

## Russian

See [../ru/KB_SOURCE_OF_TRUTH.md](../ru/KB_SOURCE_OF_TRUTH.md).
