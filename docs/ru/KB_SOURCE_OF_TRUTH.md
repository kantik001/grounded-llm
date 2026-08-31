# Source of truth для базы знаний

Postgres — **метаданные и ACL**; object storage — **версионированные blobs**; Chroma/BM25 — **пересоздаваемые индексы**.

## Архитектура

```text
Upload / connector
    → kb_documents + kb_document_versions (Postgres)
    → blobs: KB_BLOB_DIR или S3/MinIO
    → ingest → Chroma + BM25
```

## Переменные окружения

| Переменная | По умолчанию | Назначение |
|------------|--------------|------------|
| `KB_BLOB_BACKEND` | `local` | `local` или `s3` |
| `KB_BLOB_DIR` | `{DATA_DIR}/blobs` | Локальные blobs |
| `KB_S3_ENDPOINT`, `KB_S3_BUCKET`, `KB_S3_ACCESS_KEY`, `KB_S3_SECRET_KEY` | — | S3/MinIO |
| `KB_S3_PREFIX`, `KB_S3_REGION`, `KB_S3_USE_SSL` | — | Доп. настройки S3 |
| `KB_ACL_ENFORCE` | `0` | Фильтр retrieval по ACL |
| `KB_REGISTRY_SYNC` | `1` | Registry после connector sync |

MinIO: `docker compose --profile minio up -d`

## Admin API

| Метод | Путь | Назначение |
|-------|------|------------|
| GET | `/admin/kb/documents?domain_id=` | Список документов из Postgres |
| POST | `/admin/kb/index-runs?domain_id=` | Новый index run; `"activate": true` |
| POST | `/admin/upload` | blob + registry; в ответе `document_id`, `version_id` |
| DELETE | `/admin/articles?filename=` | soft-delete в registry |

## Миграция

```bash
python scripts/backfill_kb_registry.py

curl -u admin:pass -X POST "http://localhost:8080/api/admin/ingest?domain_id=default" \
  -H "Content-Type: application/json" -d '{"sync": true}'
```

## Index runs

```bash
curl -u admin:pass -X POST "http://localhost:8080/api/admin/kb/index-runs?domain_id=default" \
  -d '{"activate": true}'

curl -u admin:pass -X POST "http://localhost:8080/api/admin/ingest?domain_id=default" \
  -d '{"mode": "full"}'
```

## Connectors

Google Drive с `KB_REGISTRY_SYNC=1` пишет в registry автоматически. Иначе — `connectors/registry_sync.py` или backfill.

Полная версия: [KB_SOURCE_OF_TRUTH.md (EN)](../en/KB_SOURCE_OF_TRUTH.md)
