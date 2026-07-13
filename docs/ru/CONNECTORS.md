# Коннекторы ingest

Краткая русская версия. **Канон:** [CONNECTORS.md (EN)](../en/CONNECTORS.md).

Синхронизация документов из внешних систем в `data/{tenant}/{domain}/` перед reindex.

---

## CLI

```bash
python scripts/sync_connector.py <connector> --domain <domain_id> [options]
```

| Коннектор | Источник | Примечание |
|-----------|----------|------------|
| `local_folder` | Путь к папке | Универсальный |
| `sharepoint` | Microsoft Graph | Live API |
| `google_drive` | Google Drive API | `pip install -r api/requirements-connectors.txt` |
| `confluence` | Confluence REST | Страницы + вложения |
| `sharepoint_export` | Папка экспорта | Офлайн |
| `google_drive_export` | Takeout | Офлайн |
| `confluence_export` | Экспорт space | Офлайн |

После sync: `python scripts/reindex_rag.py`

---

## Переменные

См. `.env.example` и [CONNECTORS.md (EN)](../en/CONNECTORS.md) — SharePoint, Drive, Confluence.

---

## Связанное

- [PHASE_8.md (EN)](../en/PHASE_8.md) · [PHASE_9.md (EN)](../en/PHASE_9.md)
