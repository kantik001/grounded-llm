# Connectors (Phase 7)

Pluggable **ingest connectors** copy documents from external systems into `data/{tenant}/{domain}/` before reindex.

## Reference connector

| Connector | Module | Use |
|-----------|--------|-----|
| `local_folder` | `connectors/local_folder.py` | Mirror a directory (SharePoint export, git checkout, etc.) |

## CLI

```bash
python scripts/sync_connector.py local_folder \
  --source /path/to/docs \
  --tenant default --domain it_support

python scripts/reindex_rag.py
```

Future connectors (Phase 8+): SharePoint, Google Drive, Confluence — same `Connector` interface.

See [docs/en/CONNECTORS.md](../docs/en/CONNECTORS.md).
