# `deploy/` — reference deployment artifacts

Not application runtime code. Day-to-day local/prod compose stays at the **repo root** (`docker-compose*.yml`).

| Path | Role | Docs |
|------|------|------|
| [`helm/grounded-llm/`](./helm/grounded-llm/) | Kubernetes Helm chart (server, python, webapp, optional postgres) | [K8S_DEPLOY.md](../docs/en/K8S_DEPLOY.md) |
| [`terraform/`](./terraform/) | Cloud **reference** stacks (AWS / GCP / Azure) | [TERRAFORM.md](../docs/en/TERRAFORM.md) |

## Helm quick start

```bash
helm lint deploy/helm/grounded-llm
helm upgrade --install grounded ./deploy/helm/grounded-llm \
  --namespace grounded --create-namespace \
  -f deploy/helm/grounded-llm/values-prod.example.yaml \
  --set secrets.llmApiKey="$LLM_API_KEY" \
  --set secrets.adminSecret="$ADMIN_SECRET" \
  --set secrets.adminPassword="$ADMIN_PASSWORD" \
  --set secrets.ragServiceToken="$(openssl rand -hex 24)"
```

- Demo defaults: [`values.yaml`](./helm/grounded-llm/values.yaml) (weak postgres password, placeholder image repos — **not** for production).
- Production-shaped overlay: [`values-prod.example.yaml`](./helm/grounded-llm/values-prod.example.yaml).

### Chart vs Compose (known gaps)

| Feature | Docker Compose | Helm chart today |
|---------|----------------|------------------|
| Postgres | yes | yes (or external) |
| Redis caches | yes | **bring-your-own** — set `REDIS_URL` via env / secret |
| Python HTTP `:5000` | yes | yes |
| Python gRPC `:50051` | yes | yes (Service port `grpc`) |
| Guardrails `:50052` | optional profile | **external** — see [GUARDRAILS.md](../docs/en/GUARDRAILS.md) |

CI: job `helm-lint` runs `helm lint` + `helm template`.

## Terraform

See [`terraform/README.md`](./terraform/README.md). Modules under `terraform/{aws,gcp,azure}/reference/` are starting points, not turnkey production.

## Related

- [DEPLOY.md](../docs/en/DEPLOY.md) — Compose
- [BACKUP_RESTORE.md](../docs/en/BACKUP_RESTORE.md)
- [NETWORK_SECURITY.md](../docs/en/NETWORK_SECURITY.md)
