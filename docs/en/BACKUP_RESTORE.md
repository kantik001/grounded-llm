# Backup and restore

Grounded LLM state spans three layers. Plan backups before production go-live.

## What to back up

| Layer | Path / resource | Contains |
|-------|-----------------|----------|
| **Postgres** | `DATABASE_URL` | Users, sessions, messages, feedback, audit log, reindex jobs |
| **Chroma** | `chroma_db/` (Python PVC in K8s) | Vector embeddings index |
| **Knowledge base** | `data/{tenant}/{domain}/` | Source documents (txt, pdf, docx) |
| **Uploads** | `UPLOAD_DIR` | User image attachments |
| **Config** | `config/` | domains.json, locales, RBAC, quotas |

## Postgres

### Docker Compose

```bash
docker exec grounded_llm_postgres pg_dump -U grounded -Fc grounded > grounded-$(date +%Y%m%d).dump
```

Restore:

```bash
docker exec -i grounded_llm_postgres pg_restore -U grounded -d grounded --clean --if-exists < grounded-YYYYMMDD.dump
```

### Kubernetes

Use your cluster backup tool (Velero, CloudNativePG, RDS snapshots) on the Postgres PVC or managed instance.

## Chroma vector store

```bash
# Compose
docker cp grounded_llm_python:/app/chroma_db ./chroma_backup/

# Restore (stop python first)
docker cp ./chroma_backup/ grounded_llm_python:/app/chroma_db
docker compose restart python
```

After restore, verify with admin index stats or a retrieval eval smoke.

## Knowledge base (`data/`)

Treat as source of truth. Version in Git or object storage; reindex only when documents change.

```bash
tar czf data-backup.tar.gz data/
```

## Uploads

```bash
docker run --rm -v grounded_llm_uploads_data:/data -v "$PWD":/backup alpine \
  tar czf /backup/uploads-backup.tar.gz -C /data .
```

## Recovery order

1. Restore Postgres
2. Restore `data/` and `config/`
3. Restore Chroma **or** trigger full reindex (`POST /admin/reindex` or `FORCE_RAG_REINDEX=true`)
4. Restore uploads (optional; chat history may reference missing images)
5. Run `scripts/smoke.sh` against the API

## RPO / RTO guidance

| Tier | RPO | RTO | Approach |
|------|-----|-----|----------|
| Pilot | 24h | 4h | Daily pg_dump + weekly data tarball |
| Production | 1h | 1h | Managed DB PITR + PVC snapshots + automated reindex job |

## Related

- [K8S_DEPLOY.md](./K8S_DEPLOY.md)
- [TRUST_CENTER.md](./TRUST_CENTER.md)
