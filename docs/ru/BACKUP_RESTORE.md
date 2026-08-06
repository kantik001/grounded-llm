**Канон (EN):** [BACKUP_RESTORE.md](../en/BACKUP_RESTORE.md)

# Backup и restore

Состояние Grounded LLM — три+ слоя. Спланируйте бэкапы до prod go-live.

## Что бэкапить

| Слой | Path / resource | Содержимое |
|------|-----------------|------------|
| **Postgres** | `DATABASE_URL` | Users, sessions, messages, feedback, audit, reindex jobs, SaaS (`010`/`011`) |
| **Chroma** | `chroma_db/` (PVC в K8s) | Vector index |
| **Knowledge base** | `data/{tenant}/{domain}/` | Исходные документы |
| **Uploads** | `UPLOAD_DIR` | Картинки пользователей |
| **Config** | `config/` | domains.json, locales, RBAC, quotas |

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

## Chroma

```bash
docker cp grounded_llm_python:/app/chroma_db ./chroma_backup/

# Restore (сначала stop python)
docker cp ./chroma_backup/ grounded_llm_python:/app/chroma_db
docker compose restart python
```

Проверка: admin index stats или retrieval eval smoke.

## Knowledge base (`data/`)

Source of truth. Git или object storage; reindex только при смене документов.

```bash
tar czf data-backup.tar.gz data/
```

## Uploads

```bash
docker run --rm -v grounded_llm_uploads_data:/data -v "$PWD":/backup alpine \
  tar czf /backup/uploads-backup.tar.gz -C /data .
```

## Порядок восстановления

1. Postgres  
2. `data/` и `config/`  
3. Chroma **или** полный reindex (`POST /admin/reindex` / `FORCE_RAG_REINDEX=true`)  
4. Uploads (опц.; история может ссылаться на отсутствующие картинки)  
5. `scripts/smoke.sh` против API  

## RPO / RTO

| Tier | RPO | RTO | Подход |
|------|-----|-----|--------|
| Pilot | 24h | 4h | Daily pg_dump + weekly data tarball |
| Production | 1h | 1h | Managed DB PITR + PVC snapshots + automated reindex |

## См. также

- [K8S_DEPLOY.md](./K8S_DEPLOY.md)
- [TENANT_PURGE.md](./TENANT_PURGE.md)
- [TRUST_CENTER.md (EN)](../en/TRUST_CENTER.md)
