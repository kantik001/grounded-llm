**Канон (EN):** [K8S_DEPLOY.md](../en/K8S_DEPLOY.md)

# Kubernetes (Helm) — кратко

Развёртывание через chart `deploy/helm/grounded-llm/`. Полные детали и values — в EN-каноне и [deploy/README.md](../../deploy/README.md).

## Prerequisites

- Kubernetes 1.25+
- Helm 3.10+
- Образы (сборка из репо или GHCR)
- Storage class для Postgres и Chroma PVC

## Quick install

```bash
helm upgrade --install grounded ./deploy/helm/grounded-llm \
  --namespace grounded --create-namespace \
  --set secrets.llmApiKey="$LLM_API_KEY" \
  --set secrets.adminSecret="$ADMIN_SECRET" \
  --set secrets.ragServiceToken="$(openssl rand -hex 24)" \
  --set secrets.adminPassword="$ADMIN_PASSWORD"
```

## Архитектура

```text
Ingress (optional)
    ├── webapp (nginx)  → UI + /api proxy
    └── server (Go :8080)
            ├── postgres (StatefulSet)
            ├── python (RAG HTTP :5000 + gRPC :50051) → chroma PVC
            ├── redis (BYO — в chart пока нет; задайте REDIS_URL)
            └── guardrails :50052 (внешний, опц. — см. GUARDRAILS.md)
```

**Chart lag:** Redis в chart **не** поставляется. Python gRPC `:50051` — на Service python. `LLM_PROVIDER` / `LLM_BASE_URL` — как в Compose.

## Health probes

| Service | Liveness | Readiness |
|---------|----------|-----------|
| Go `:8080` | `GET /health` | `GET /ready` (Postgres + Python RAG) |
| Python | `GET /health` | `GET /ready` (+ `X-RAG-Service-Token`); startup ~6 мин |
| Postgres | `pg_isready` | `pg_isready` |
| Webapp | `GET /` | `GET /` |

Один и тот же `RAG_SERVICE_TOKEN` на Go и Python.

## Production checklist

1. Secrets — External Secrets / sealed; не коммитить  
2. KB — mount `data/` + `config/` до первого index  
3. Ingress TLS + ограничение admin на edge  
4. Retention: `messageRetentionDays` / `sessionRetentionDays`  
5. Backups — [BACKUP_RESTORE.md](./BACKUP_RESTORE.md)  
6. Scraping `GET /metrics` (Go); логи с `X-Request-ID`  
7. NetworkPolicy — см. [NETWORK_SECURITY.md](./NETWORK_SECURITY.md)  

Values: `values.yaml` (demo), `values-prod.example.yaml` (prod-shaped). Внешний Postgres: `postgres.enabled: false` + `DATABASE_URL`.

## См. также

- [DEPLOY.md](./DEPLOY.md) — Compose  
- [TERRAFORM.md (EN)](../en/TERRAFORM.md)  
- [COMPATIBILITY.md](./COMPATIBILITY.md)  
