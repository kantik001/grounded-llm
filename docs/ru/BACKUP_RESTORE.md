**Канон (EN):** [BACKUP_RESTORE.md](../en/BACKUP_RESTORE.md)

# Backup и restore

Состояние Grounded LLM — три+ слоя. Спланируйте бэкапы до prod go-live.

## Что бэкапить

| Слой | Path / resource | Содержимое | SoT? |
|------|-----------------|------------|------|
| **Postgres** | `DATABASE_URL` | Users, sessions, **kb_documents**, ingest jobs, … | **Да** (metadata) |
| **KB blobs** | `KB_BLOB_DIR` или S3 | Версии документов | **Да** (content) |
| **Legacy `data/`** | `data/{tenant}/{domain}/` | Dual-write копия | Опционально |
| **Chroma / BM25** | `chroma_db/`, sparse | Индексы | **Нет** — rebuild через ingest |
| **Uploads** | `UPLOAD_DIR` | Картинки чата | Да |
| **Config** | `config/` | domains, locales | Да |

См. [KB_SOURCE_OF_TRUTH.md](./KB_SOURCE_OF_TRUTH.md).

При `VECTOR_STORE=qdrant` / `pgvector` — бэкапьте соответствующий store (снимки Qdrant / таблицы Postgres). Sparse BM25: `SPARSE_INDEX_DIR` / `sparse_index/`.

## Postgres

### Docker Compose

```bash
docker exec grounded_llm_postgres pg_dump -U grounded -Fc grounded > grounded-$(date +%Y%m%d).dump
```

Restore:

```bash
docker exec -i grounded_llm_postgres pg_restore -U grounded -d grounded --clean --if-exists < grounded-YYYYMMDD.dump
```

### Smoke (CI / local)

```bash
PGPASSWORD=grounded bash scripts/backup_postgres_smoke.sh
# или
make backup-smoke
```

### Kubernetes

Velero, CloudNativePG, RDS snapshots — на PVC Postgres или managed instance.

## Chroma / sparse (disposable)

Индексы можно пересобрать из Postgres + blobs. Бэкап Chroma опционален.

```bash
curl -u admin:pass -X POST "http://localhost:8080/api/admin/ingest?domain_id=default" \
  -d '{"mode": "full", "sync": true}'
```

## Knowledge base blobs

**Primary:** Postgres + blob store. Legacy `data/` — опционально.

```bash
tar czf kb-blobs-backup.tar.gz data/blobs/
tar czf data-backup.tar.gz data/
```

## Uploads

```bash
docker run --rm -v grounded_llm_uploads_data:/data -v "$PWD":/backup alpine \
  tar czf /backup/uploads-backup.tar.gz -C /data .
```

## Порядок восстановления

1. Postgres (`kb_*`, ACL, index runs)  
2. Blob store + `config/`  
3. Опционально `data/`  
4. Full ingest или restore Chroma  
5. Uploads (опц.)  
6. `scripts/smoke.sh`  

## RPO / RTO

| Tier | RPO | RTO | Подход |
|------|-----|-----|--------|
| Pilot | 24h | 4h | Daily pg_dump + weekly data tarball |
| Production | 1h | 1h | Managed DB PITR + PVC snapshots + automated reindex |

## См. также

- [K8S_DEPLOY.md](./K8S_DEPLOY.md)
- [TENANT_PURGE.md](./TENANT_PURGE.md)
- [TRUST_CENTER.md (EN)](../en/TRUST_CENTER.md)
