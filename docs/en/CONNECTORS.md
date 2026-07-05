# Ingest connectors

Connectors sync documents from external systems into `data/{tenant}/{domain}/` before RAG reindex.

---

## Architecture

```text
External source  →  Connector.sync()  →  data/{tenant}/{domain}/
                                              ↓
                                    python scripts/reindex_rag.py
```

Supported file types: `.txt`, `.pdf`, `.docx`.

---

## CLI

```bash
python scripts/sync_connector.py <connector> --domain <domain_id> [options]
```

| Connector | `--source` | Notes |
|-----------|------------|-------|
| `local_folder` | Required path | Generic folder mirror |
| `sharepoint_export` | Export folder | SharePoint / OneDrive synced folder |
| `google_drive_export` | Takeout folder | Google Drive export |
| `confluence_export` | Space export | Confluence PDF/attachments tree |
| `sharepoint` | Optional subfolder | **Live** Microsoft Graph (env vars) |

Examples:

```bash
python scripts/sync_connector.py local_folder --source ./packs/hr/data --domain default
python scripts/sync_connector.py sharepoint_export --source /data/sp-export --domain legal_faq
python scripts/sync_connector.py sharepoint --domain it_support --dry-run
```

Then:

```bash
python scripts/reindex_rag.py
python scripts/run_rag_eval.py --suite it_support
```

---

## SharePoint Graph (live)

Set environment variables (never commit secrets):

| Variable | Description |
|----------|-------------|
| `SHAREPOINT_TENANT_ID` | Azure AD tenant |
| `SHAREPOINT_CLIENT_ID` | App registration client id |
| `SHAREPOINT_CLIENT_SECRET` | Client secret |
| `SHAREPOINT_DRIVE_ID` | Graph drive id |
| `SHAREPOINT_FOLDER_PATH` | Optional subfolder inside drive |

App registration needs `Sites.Read.All` / `Files.Read.All` application permissions.

---

## Planned (Phase 9+)

- Google Drive API connector (service account)
- Confluence REST API connector
- Scheduled sync via admin job / cron

---

## Related

- [connectors/README.md](../../connectors/README.md)
- [packs/README.md](../../packs/README.md)
