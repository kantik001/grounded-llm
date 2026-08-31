# Ingest connectors

Connectors download documents into a **staging directory**, register them in the **KB registry** (Postgres + blobs), then index via **ingest** or **reindex**.

- Ingest: [docs/en/INGESTION.md](../docs/en/INGESTION.md)
- Registry / SoT: [docs/en/KB_SOURCE_OF_TRUTH.md](../docs/en/KB_SOURCE_OF_TRUTH.md)
- Reindex fallback: `python scripts/reindex_rag.py`

## Connectors

| Name | Module | Mode |
|------|--------|------|
| `local_folder` | `local_folder.py` | Folder mirror |
| `sharepoint_export` | `sharepoint_export.py` | Offline SharePoint export |
| `google_drive_export` | `google_drive_export.py` | Drive Takeout folder |
| `confluence_export` | `confluence_export.py` | Confluence space export |
| `sharepoint` | `sharepoint.py` | Live Microsoft Graph |
| `google_drive` | `google_drive.py` | Live Google Drive API (+ registry sync) |
| `confluence` | `confluence.py` | Live Confluence REST |

Registry helper: [`registry_sync.py`](./registry_sync.py) — `register_synced_file`, `register_synced_tree`.

| Env | Default | Purpose |
|-----|---------|---------|
| `KB_REGISTRY_SYNC` | `1` | Auto-register after Google Drive sync |

Registry (names + factories): [`registry.py`](./registry.py) — `build_connector(name, source)`.

Optional Google Drive deps:

```bash
pip install -r connectors/requirements.txt
# alias: api/requirements-connectors.txt
```

## CLI

```bash
python scripts/sync_connector.py <connector> --domain <id> [--source PATH] [--dry-run]
```

`sync_connector.py` registers synced files automatically, then run ingest:

```bash
curl -u admin:pass -X POST "http://localhost:8080/api/admin/ingest?domain_id=<id>" \
  -H "Content-Type: application/json" -d '{"sync": false}'
```

Exit code `0` if `SyncResult.ok` (no errors); `1` on setup failure or sync errors.

## Tests / CI

Unit tests: `tests/test_connectors.py`, `test_export_connectors.py`, `test_*_connector.py` — run in CI job **`python-test`**.

```bash
pytest tests/test_connectors.py tests/test_export_connectors.py tests/test_google_drive_connector.py tests/test_confluence_connector.py -v
```

See [docs/en/CONNECTORS.md](../docs/en/CONNECTORS.md).
