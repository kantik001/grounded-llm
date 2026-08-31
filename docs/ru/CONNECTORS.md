# Коннекторы ingest

Коннекторы скачивают документы во **staging**, регистрируют их в **KB registry** (Postgres + blobs), затем индексируют через **ingest** или **reindex**.

- Registry: [KB_SOURCE_OF_TRUTH.md](./KB_SOURCE_OF_TRUTH.md)
- Рекомендуется: `POST /admin/ingest` — [INGESTION.md](./INGESTION.md)
- Fallback: `python scripts/reindex_rag.py`

```text
External source  →  Connector.sync()  →  connector_staging/{tenant}/{domain}/
                         ↓
                    register_synced_tree() → kb_documents + blobs
                                              ↓
                         POST /admin/ingest  или  reindex_rag.py
```

Миграция старых файлов из `data/`: `python scripts/backfill_kb_registry.py`.

Подробности: [../en/CONNECTORS.md](../en/CONNECTORS.md) (EN).
