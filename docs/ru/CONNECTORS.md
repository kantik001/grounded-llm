# Коннекторы ingest

Краткая русская версия. **Канон:** [CONNECTORS.md (EN)](../en/CONNECTORS.md).

Синхронизация в `data/{tenant}/{domain}/`, registry (Postgres + blobs), затем **ingest** или **reindex**.

См. [KB_SOURCE_OF_TRUTH.md](./KB_SOURCE_OF_TRUTH.md) · [INGESTION.md](./INGESTION.md).

`KB_REGISTRY_SYNC=1` — Google Drive автоматически регистрирует файлы в registry. Иначе: `python scripts/backfill_kb_registry.py`.

---

## CLI

```bash
python scripts/sync_connector.py <connector> --domain <domain_id> [options]
```

| Коннектор | Источник | Примечание |
|-----------|----------|------------|
| `local_folder` | Путь к папке | Универсальный |
| `sharepoint` | Microsoft Graph | Live API |
| `google_drive` | Google Drive API | `pip install -r connectors/requirements.txt` |
| `confluence` | Confluence REST | Страницы + вложения |
| `sharepoint_export` | Папка экспорта | Офлайн |
| `google_drive_export` | Takeout | Офлайн |
| `confluence_export` | Экспорт space | Офлайн |

После sync: `POST /admin/ingest` (прод) или `python scripts/reindex_rag.py` (dev/CI)

---

## Переменные

См. `.env.example` и [CONNECTORS.md (EN)](../en/CONNECTORS.md) — SharePoint, Drive, Confluence.

---

## Связанное

- [PHASE_8.md (EN)](../en/PHASE_8.md) · [PHASE_9.md (EN)](../en/PHASE_9.md)
