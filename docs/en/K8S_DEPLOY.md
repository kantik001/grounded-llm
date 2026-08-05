# Kubernetes deployment (Helm)

Deploy Grounded LLM on Kubernetes using the Helm chart under `deploy/helm/grounded-llm/`.

## Prerequisites

- Kubernetes 1.25+
- Helm 3.10+
- Container images (build from repo or use GHCR release tags)
- Persistent storage class for Postgres and Chroma PVCs

## Quick install

```bash
helm upgrade --install grounded ./deploy/helm/grounded-llm \
  --namespace grounded --create-namespace \
  --set secrets.llmApiKey="$LLM_API_KEY" \
  --set secrets.adminSecret="$ADMIN_SECRET" \
  --set secrets.ragServiceToken="$(openssl rand -hex 24)" \
  --set secrets.adminPassword="$ADMIN_PASSWORD"
```

## Architecture

```text
Ingress (optional)
    ├── webapp (nginx)  → static UI + /api proxy
    └── server (Go :8080)
            ├── postgres (StatefulSet)
            ├── python (RAG HTTP :5000 + gRPC :50051) → chroma PVC
            ├── redis (optional — bring-your-own; Compose always includes it)
            └── guardrails :50052 (optional external — not in Helm yet; see GUARDRAILS.md)
```

> **Chart lag:** Redis is **not** shipped in the chart — bring Redis and set `REDIS_URL` (or disable caches). Python gRPC `:50051` is exposed on the python Service. Set `LLM_PROVIDER` / `LLM_BASE_URL` via secrets/env like Compose. See [LLM_PROVIDERS.md](./LLM_PROVIDERS.md) and [deploy/README.md](../../deploy/README.md).

## Health probes

| Service | Liveness | Readiness | Startup |
|---------|----------|-----------|---------|
| Go server | `GET /health` | `GET /ready` (Postgres + Python RAG) | — |
| Python RAG | `GET /health` | `GET /ready` (+ `X-RAG-Service-Token`) | `GET /health` (covers model/index warm-up; ~6 min budget) |
| Postgres | `pg_isready` | `pg_isready` | — |
| Webapp | `GET /` | `GET /` | — |

Tune timings under `*.probes` in `values.yaml` (timeouts, `failureThreshold`, Python `startupProbe`).

Set the same `RAG_SERVICE_TOKEN` on Go server and Python service. Go sends `X-RAG-Service-Token` on internal calls.

## Production checklist

1. **Secrets** — use External Secrets Operator or sealed secrets; never commit real values.
2. **Knowledge base** — mount `data/` and `config/` via ConfigMap/CSI or sync from object storage before first index.
3. **Ingress TLS** — enable `ingress.tls` and restrict admin routes at the edge.
4. **Retention** — set `retention.messageRetentionDays` / `sessionRetentionDays` per policy.
5. **Backups** — schedule [BACKUP_RESTORE.md](./BACKUP_RESTORE.md) for Postgres, Chroma PVC, and uploads.
6. **Observability** — scrape `GET /metrics` from the Go server; ship logs with `X-Request-ID` correlation.

## Customize values

- Demo: `deploy/helm/grounded-llm/values.yaml`
- Production-shaped example: `deploy/helm/grounded-llm/values-prod.example.yaml`
- Map: [deploy/README.md](../../deploy/README.md)

For external managed Postgres, set `postgres.enabled: false` and point `DATABASE_URL` via a custom values overlay (patch server deployment env).

## Related

- [DEPLOY.md](./DEPLOY.md) — Docker Compose
- [BACKUP_RESTORE.md](./BACKUP_RESTORE.md)
- [NETWORK_SECURITY.md](./NETWORK_SECURITY.md)
- [TRUST_CENTER.md](./TRUST_CENTER.md)
