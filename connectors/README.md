# Ingest connectors

Connectors sync documents from external systems into `data/{tenant}/{domain}/` before RAG reindex.

## Connectors

| Name | Module | Mode |
|------|--------|------|
| `local_folder` | `local_folder.py` | Folder mirror |
| `sharepoint_export` | `sharepoint_export.py` | Offline SharePoint export |
| `google_drive_export` | `google_drive_export.py` | Drive Takeout folder |
| `confluence_export` | `confluence_export.py` | Confluence space export |
| `sharepoint` | `sharepoint.py` | Live Microsoft Graph |
| `google_drive` | `google_drive.py` | Live Google Drive API |
| `confluence` | `confluence.py` | Live Confluence REST |

Optional deps for Google Drive: `pip install -r api/requirements-connectors.txt`

See [docs/en/CONNECTORS.md](../docs/en/CONNECTORS.md).
