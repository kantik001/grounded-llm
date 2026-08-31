# Backup and restore

Grounded LLM state spans three layers. Plan backups before production go-live.

## What to back up

| Layer | Path / resource | Contains | SoT? |
|-------|-----------------|----------|------|
| **Postgres** | `DATABASE_URL` | Users, sessions, messages, audit, **kb_documents**, ingest jobs | **Yes** (metadata) |
| **KB blobs** | `KB_BLOB_DIR` or S3 (`KB_S3_*`) | Versioned document bytes | **Yes** (content) |
| **Chroma / BM25** | `chroma_db/`, sparse `.pkl` | Vector + sparse indexes | **No** — rebuild via ingest |
| **Uploads** | `UPLOAD_DIR` | User image attachments | Yes (chat images) |
| **Config** | `config/` | domains.json, locales, RBAC, quotas | Yes |

See [KB_SOURCE_OF_TRUTH.md](./KB_SOURCE_OF_TRUTH.md) for the enterprise document model.

## Postgres

### Docker Compose

```bash
docker exec grounded_llm_postgres pg_dump -U grounded -Fc grounded > grounded-$(date +%Y%m%d).dump
```

Restore:

```bash
docker exec -i grounded_llm_postgres pg_restore -U grounded -d grounded --clean --if-exists < grounded-YYYYMMDD.dump
```

### Smoke test (CI / local)

Applies migrations, seeds a marker row, `pg_dump` → `pg_restore` into a throwaway DB, verifies the row:

```bash
# Compose Postgres on localhost:5432
PGPASSWORD=grounded bash scripts/backup_postgres_smoke.sh

# Or via Makefile (same defaults)
make backup-smoke
```

### Kubernetes

Use your cluster backup tool (Velero, CloudNativePG, RDS snapshots) on the Postgres PVC or managed instance.

## Chroma / sparse indexes (disposable)

Indexes can be rebuilt from Postgres + blobs. Backing up Chroma is optional (faster recovery); skipping it is valid if you can run a full ingest after restore.

```bash
# Optional — speeds recovery
docker cp grounded_llm_python:/app/chroma_db ./chroma_backup/
```

After restore without Chroma backup:

```bash
curl -u admin:pass -X POST "http://localhost:8080/api/admin/ingest?domain_id=default" \
  -H "Content-Type: application/json" -d '{"mode": "full", "sync": true}'
```

Legacy one-shot: `POST /admin/reindex` or `FORCE_RAG_REINDEX=true`.

## Knowledge base blobs

**Primary:** Postgres (`kb_*` tables) + blob store.

```bash
# Local blobs
tar czf kb-blobs-backup.tar.gz -C "$(dirname "$KB_BLOB_DIR")" "$(basename "$KB_BLOB_DIR")"
```

For S3/MinIO: use bucket versioning and provider backup policies (`KB_S3_BUCKET`).

## Uploads

```bash
docker run --rm -v grounded_llm_uploads_data:/data -v "$PWD":/backup alpine \
  tar czf /backup/uploads-backup.tar.gz -C /data .
```

## Recovery order

1. Restore Postgres (includes `kb_documents`, ACL, index run metadata)
2. Restore blob store (`KB_BLOB_DIR` or S3) and `config/`
3. Rebuild indexes: `POST /admin/ingest` with `"mode": "full"` (or restore Chroma backup)
4. Restore uploads (optional; chat history may reference missing images)
6. Run `scripts/smoke.sh` against the API

## RPO / RTO guidance

| Tier | RPO | RTO | Approach |
|------|-----|-----|----------|
| Pilot | 24h | 4h | Daily pg_dump + weekly data tarball |
| Production | 1h | 1h | Managed DB PITR + S3 bucket versioning + full ingest after restore |

## Related

- [K8S_DEPLOY.md](./K8S_DEPLOY.md)
- [TRUST_CENTER.md](./TRUST_CENTER.md)
