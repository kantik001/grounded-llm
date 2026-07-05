# Ingest connectors

Connectors sync documents from external systems into `data/{tenant}/{domain}/` before RAG reindex.

Phase 7 ships a **reference interface** and one production-shaped connector. Enterprise connectors (SharePoint, Google Drive, Confluence) follow the same pattern in Phase 8+.

---

## Architecture

```text
External source  →  Connector.sync()  →  data/{tenant}/{domain}/
                                              ↓
                                    python scripts/reindex_rag.py
```

All connectors implement `connectors.base.Connector`:

| Method | Purpose |
|--------|---------|
| `sync(target_dir, dry_run=False)` | Copy supported files into KB directory |

Supported file types: `.txt`, `.pdf`, `.docx` (same as core pipeline).

---

## Reference: local_folder

Mirror a directory (e.g. SharePoint export, git checkout):

```bash
python scripts/sync_connector.py local_folder \
  --source /path/to/export \
  --tenant default \
  --domain it_support

python scripts/reindex_rag.py
python scripts/run_rag_eval.py --suite it_support
```

Dry run:

```bash
python scripts/sync_connector.py local_folder --source ./packs/hr/data --domain default --dry-run
```

---

## Planned connectors (Phase 8+)

| Connector | Source |
|-----------|--------|
| `sharepoint` | Microsoft Graph / SharePoint Online |
| `google_drive` | Google Drive API |
| `confluence` | Atlassian REST API |

Each connector should:

1. Authenticate via env / secret (never commit tokens)
2. Write into tenant/domain data path
3. Log sync summary (`SyncResult`)
4. Trigger async reindex via admin API or `reindex_rag.py`

---

## Related

- [packs/README.md](../../packs/README.md)
- [DEPLOY.md](./DEPLOY.md)
