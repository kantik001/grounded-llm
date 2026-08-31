# Ingest connectors

Connectors download documents into a **staging directory**, register them in the **KB registry** (Postgres + blobs), then index via **ingest** or **reindex**.

- **Registry:** [KB_SOURCE_OF_TRUTH.md](./KB_SOURCE_OF_TRUTH.md)
- **Recommended:** `POST /admin/ingest` after sync — [INGESTION.md](./INGESTION.md)
- **Fallback:** `python scripts/reindex_rag.py` or `POST /admin/reindex`

---

## Architecture

```text
External source  →  Connector.sync()  →  connector_staging/{tenant}/{domain}/
                         ↓
                    register_synced_tree() → kb_documents + blobs
                                              ↓
                         POST /admin/ingest  (async)  or  reindex_rag.py (sync)
```

| Variable | Default | Purpose |
|----------|---------|---------|
| `KB_REGISTRY_SYNC` | `1` | Auto-register synced files (Google Drive) |
| `KB_BLOB_BACKEND` | `local` | Blob storage backend |

One-time migration from old `data/` trees: `python scripts/backfill_kb_registry.py`.

Supported file types: `.txt`, `.pdf`, `.docx` (+ Google Docs exported as `.txt`).

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
| `sharepoint` | Optional subfolder | Live Microsoft Graph |
| `google_drive` | — | Live Google Drive API + registry sync |
| `confluence` | — | Live Confluence REST |

Examples:

```bash
python scripts/sync_connector.py confluence --domain it_support --dry-run
pip install -r connectors/requirements.txt
python scripts/sync_connector.py google_drive --domain default --dry-run
```

Then index (pick one):

```bash
# Async ingest (production)
curl -u admin:pass -X POST "http://localhost:8080/api/admin/ingest?domain_id=it_support" \
  -H "Content-Type: application/json" -d '{"sync": false}'

# Or sync reindex (dev/CI)
python scripts/reindex_rag.py
python scripts/run_rag_eval.py --suite it_support
```

---

## SharePoint Graph

| Variable | Description |
|----------|-------------|
| `SHAREPOINT_TENANT_ID` | Azure AD tenant |
| `SHAREPOINT_CLIENT_ID` | App client id |
| `SHAREPOINT_CLIENT_SECRET` | Client secret |
| `SHAREPOINT_DRIVE_ID` | Graph drive id |
| `SHAREPOINT_FOLDER_PATH` | Optional subfolder |

---

## Google Drive API

| Variable | Description |
|----------|-------------|
| `GOOGLE_APPLICATION_CREDENTIALS` | Service account JSON path |
| `GOOGLE_DRIVE_FOLDER_ID` | Shared folder id |
| `GOOGLE_DRIVE_IMPERSONATE_USER` | Optional domain-wide delegation |

Install: `pip install -r connectors/requirements.txt`

Share the target folder with the service account email. With `KB_REGISTRY_SYNC=1`, synced files are registered in Postgres + blob store automatically.

---

## Confluence REST

| Variable | Description |
|----------|-------------|
| `CONFLUENCE_BASE_URL` | e.g. `https://your.atlassian.net/wiki` |
| `CONFLUENCE_EMAIL` | Atlassian account email |
| `CONFLUENCE_API_TOKEN` | API token |
| `CONFLUENCE_SPACE_KEY` | Space to export |

Pages are saved as `.txt` (HTML stripped); attachments copied when supported.

---

## Related

- [connectors/README.md](../../connectors/README.md)
- [KB_SOURCE_OF_TRUTH.md](./KB_SOURCE_OF_TRUTH.md)
- [LAUNCH.md](./LAUNCH.md)
