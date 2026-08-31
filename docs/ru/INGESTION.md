# Ingestion pipeline

Асинхронная индексация KB: **parse → embed → index**. Состояние job — в Postgres, очередь задач — в Redis.

См. также: [KB_SOURCE_OF_TRUTH.md](./KB_SOURCE_OF_TRUTH.md) · [CONNECTORS.md](./CONNECTORS.md) · `POST /admin/reindex` как fallback.

## Архитектура

**Source of truth (prod):** Postgres (`kb_documents`) + blob store (`KB_BLOB_DIR` или S3). Каталог `data/{tenant}/{domain}/` по-прежнему получает dual-write при upload и используется как fallback, если registry пуст.

```
Upload / connector
    → kb_documents + kb_document_versions (Postgres)
    → blobs (local / S3)
    → data/{tenant}/{domain}/          (legacy)
        ↓
POST /admin/ingest (Go) → ingest_jobs
        ↓
Python worker → Redis → discover (Postgres → fallback data/)
    → parse → staging → embed → Chroma → BM25
```

Путь retrieve (`/rag/context`) не блокируется. `KB_ACL_ENFORCE=1` — фильтр по ACL.

## API

```http
POST /admin/ingest?domain_id=hr
{"files": [], "mode": "incremental", "sync": false}

GET /admin/ingest/status?job_id=42
GET /admin/kb/documents?domain_id=hr
POST /admin/kb/index-runs?domain_id=hr
{"activate": true}
```

Пустой `files` = все активные документы домена (из registry или `data/`).

## Переменные окружения

| Переменная | Назначение |
|------------|------------|
| `INGEST_ASYNC`, `INGEST_STAGING_DIR`, `DATABASE_URL`, `REDIS_URL` | Ingest pipeline |
| `KB_BLOB_BACKEND`, `KB_BLOB_DIR`, `KB_S3_*` | Blob store — см. [KB_SOURCE_OF_TRUTH.md](./KB_SOURCE_OF_TRUTH.md) |
| `KB_ACL_ENFORCE` | ACL на retrieval |
| `KB_REGISTRY_SYNC` | Auto registry после connector sync |

## Локально

```bash
python scripts/backfill_kb_registry.py

curl -X POST "http://localhost:8080/admin/ingest?domain_id=default" \
  -H "X-Admin-Secret: $ADMIN_SECRET" -d '{"sync": true}'

docker compose up -d ingest-worker
```

`POST /admin/reindex` сохранён для dev/CI.
