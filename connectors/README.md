# Connectors (Phase 7–8)

Pluggable **ingest connectors** copy documents into `data/{tenant}/{domain}/` before reindex.

## Connectors

| Name | Module | Mode |
|------|--------|------|
| `local_folder` | `local_folder.py` | Folder mirror |
| `sharepoint_export` | `sharepoint_export.py` | Offline SharePoint export |
| `google_drive_export` | `google_drive_export.py` | Drive Takeout folder |
| `confluence_export` | `confluence_export.py` | Confluence space export |
| `sharepoint` | `sharepoint.py` | Live Microsoft Graph |

```bash
python scripts/sync_connector.py sharepoint_export --source /path --domain it_support --dry-run
```

See [docs/en/CONNECTORS.md](../docs/en/CONNECTORS.md).
