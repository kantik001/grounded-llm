# Deployment

Guide for running a **new project** on the Grounded LLM scaffold.  
Architecture: [ARCHITECTURE.md](./ARCHITECTURE.md) · Local LLM / caches: [LLM_PROVIDERS.md](./LLM_PROVIDERS.md).

---

## Quick start (Docker)

```bash
cp .env.example .env
# Cloud LLM: set LLM_API_KEY (OpenAI-compatible)
# Local LLM (CPU): LLM_PROVIDER=ollama  →  docker compose --profile ollama up -d --build
# Local LLM (GPU): LLM_PROVIDER=vllm    →  docker compose --profile vllm up -d --build
# Browser without Telegram: TELEGRAM_AUTH_DISABLED=true

docker compose up -d --build
```

| Service | URL / port |
|---------|------------|
| Web App | http://localhost/ |
| Go API | http://localhost:8080/health |
| Go metrics | http://localhost:8080/metrics |
| Python RAG (HTTP) | http://127.0.0.1:5000/health (loopback only) |
| gRPC Retriever | `localhost:50051` (`grounded.rag.v1.Retriever`) |
| Guardrails (optional) | `localhost:50052` — [GUARDRAILS.md](./GUARDRAILS.md) |
| Redis | `localhost:6379` (embedding + response cache) |

After adding documents under `data/`:

```bash
# Production (async per-file pipeline)
curl -u admin:pass -X POST "http://localhost:8080/api/admin/ingest?domain_id=default" \
  -H "Content-Type: application/json" -d '{"sync": false}'
docker compose up -d ingest-worker

# Dev / CI (sync fallback)
python scripts/reindex_rag.py
# or POST /admin/reindex
```

See [INGESTION.md](./INGESTION.md).

Supported KB formats: **`.txt`**, **`.pdf`**, **`.docx`**.

**Production-shaped Compose** (required secrets, no public Python/Postgres/gRPC ports):

```bash
# GROUNDED_ENV=production, ADMIN_PASSWORD, ADMIN_SECRET, RAG_SERVICE_TOKEN,
# POSTGRES_PASSWORD, DATABASE_URL, CORS_ALLOWED_ORIGINS, LLM_API_KEY (or local provider)
make up-prod
```

---

## Config without rebuild

The `./config` directory is mounted into containers as `/config` (read-only).

| Variable | File / path |
|----------|-------------|
| `DOMAINS_CONFIG_PATH` | `domains.json` |
| `LOCALES_ROOT` | `config/locales` (`ru/`, `en/`) |
| `DEFAULT_LOCALE` | `en` or `ru` (code default: **`en`**) |
| `DEFAULT_TENANT_ID` | default tenant for KB paths |
| `LLM_PROVIDER` | `openai` (default) \| `ollama` \| `vllm` |
| `REDIS_URL` | e.g. `redis://redis:6379/0` |
| `API_KEYS` or `API_KEYS_FILE` | integrator API keys |

**Reload Go without restart:**

```bash
docker compose kill -s HUP server
```

Or set `CONFIG_RELOAD_INTERVAL_SEC=300` in `.env`.

Python `rag/domains_config.py` reloads `domains.json` when mtime changes.

---

## Local development (without Docker)

1. Postgres + Redis + `.env` (`DATABASE_URL`, optional `REDIS_URL`).
2. `cd server && go run ./cmd/server`
3. Python RAG: from repo root, prefer the same entrypoint as Docker:
   ```bash
   # HTTP only (dev): python -m flask --app api.http.app run -p 5000
   # Or full stack (Gunicorn + gRPC): sh api/entrypoint.sh
   ```
4. Web: nginx or `webapp/` + `TELEGRAM_AUTH_DISABLED=true`, API on `:8080`.

---

## Eval after KB changes

```bash
pip install requests
set PYTHON_RAG_URL=http://localhost:5000/rag/context
python scripts/run_rag_eval.py --suite default
make eval-retrieval
```

Results: `eval/results/YYYYMMDD_HHMMSS.json`.

Run after: reindex, locale prompt changes, `LLM_MODEL` / embedding model change.

---

## New customer: domain pack

### 1. Repository

```bash
git clone <url> client-assistant
cd client-assistant
```

### 2. Domain pack

| Action | Path |
|--------|------|
| KB documents | `data/{tenant_id}/{domain_id}/` (`.txt`, `.pdf`, `.docx`) |
| Domain catalog | `config/domains.json` |
| Prompts & few-shot | `config/locales/ru/`, `config/locales/en/` |
| UI branding | locale `branding.json`; customize `webapp/` if needed |
| Eval questions | `eval/rag_{domain}_baseline.jsonl` |

Scaffold: `python scripts/init_pack.py install <pack_id>` (preferred) or `scripts/init_domain.ps1` / `init_domain.sh`.

### 3. Index and verify

```bash
python scripts/reindex_rag.py
python scripts/run_rag_eval.py --suite default
```

### 4. Secrets

`.env`: `LLM_API_KEY` (or local `LLM_PROVIDER`), `REDIS_URL`, `DATABASE_URL`, `CORS`, Telegram, `ADMIN_PASSWORD`, `ADMIN_SECRET`, `RAG_SERVICE_TOKEN` (prod), optional `API_KEYS`.

### 5. Pilot metrics

Verify pass rate, “not in materials” rate, thumbs up/down, latency p95 / TTFT.  
Prometheus: `GET /metrics` (`llm_tokens_*`, `llm_ttft_*`, cache counters).

---

## Smoke

```bash
make smoke
# TELEGRAM_AUTH_DISABLED=true, server on :8080
```

---

## Do not copy to a new instance

- volume `chroma_data` (recreated by reindex).
- `postgres_data` / production sessions.
- Redis cache volume data (optional — caches are ephemeral by design).
- `.env` secrets — only `.env.example` as template.

---

## Optional modules

**Vision / CV** — separate domain pack, not part of platform core.

**Hosted SaaS signup** — disabled by default. To enable self-serve tenant creation + Stripe billing, see [SAAS.md](./SAAS.md) and [BILLING.md](./BILLING.md). Not required for on-prem pilots.

**Kubernetes:** see [K8S_DEPLOY.md](./K8S_DEPLOY.md). Helm chart may lag Compose for Redis/gRPC — bring Redis via your cluster or extend the chart.
