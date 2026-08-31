# Ingestion pipeline

Асинхронная индексация KB: **parse → embed → index**. Состояние job — в Postgres, очередь задач — в Redis.

См. также: [CONNECTORS.md](./CONNECTORS.md) · `POST /admin/reindex` как fallback полной синхронизации.

## Архитектура

```
Upload / connector → data/{tenant}/{domain}/
        ↓
POST /admin/ingest (Go) → ingest_jobs
        ↓
Python worker → Redis → parse → staging → embed → Chroma → BM25
```

Путь retrieve (`/rag/context`) не блокируется на ingestion.

## Внедрение по неделям

| Неделя | Что сделано |
|--------|-------------|
| 1 | Таблица jobs, API enqueue, sync-режим (`sync=true`) |
| 2 | Redis + parse worker + staging чанков |
| 3 | Embed worker + upsert в Chroma |
| 4 | Status API, retry, DLQ, метрики |
| 5+ | Масштабирование воркеров, auto-enqueue из connectors |

## API

```http
POST /admin/ingest?domain_id=hr
{"files": [], "mode": "incremental", "sync": false}

GET /admin/ingest/status?job_id=42
```

Пустой `files` = все поддерживаемые файлы в домене.

## Переменные окружения

- `INGEST_ASYNC=1` — очередь Redis (0 = обработка в процессе API)
- `INGEST_STAGING_DIR` — каталог staging
- `DATABASE_URL`, `REDIS_URL` — обязательны для async

## Локально

```bash
# Неделя 1 — без воркера
curl -X POST "http://localhost:8080/admin/ingest?domain_id=default" \
  -H "X-Admin-Secret: $ADMIN_SECRET" -d '{"sync": true}'

# Неделя 2+
docker compose up -d ingest-worker
```

`POST /admin/reindex` сохранён для dev/CI и полного reindex.
