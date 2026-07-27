# `api/app.py` — Python RAG API

**Source:** `api/app.py`, `api/grpc_retriever.py`, `api/entrypoint.sh`  
**Language:** Python (Flask app module; production HTTP via **Gunicorn**)  
**Related:** `rag/retrieval.py`, `rag/embedding_cache.py`, `rag/vector_store.py`, `rag/domains_config.py`  
**Called by:** Go server (`PYTHON_RAG_URL`); agents via gRPC `:50051`

---

## Purpose

Separate **Python service** (Compose service **`python`**):

| Surface | Port | Role |
|---------|------|------|
| HTTP (Gunicorn) | **5000** | Retrieval + admin reindex |
| gRPC Retriever | **50051** | Agent-friendly `Retrieve` RPC |

| Endpoint | Purpose |
|----------|---------|
| `POST /rag/context` | Retrieval: article fragments for a question (no LLM) |
| `GET /domains` | Domain catalog from `config/domains.json` |
| `GET /health` | Liveness |
| `GET /ready` | Readiness (requires `RAG_SERVICE_TOKEN` when set) |
| `GET /metrics` | Embedding cache hit/miss counters |
| `GET /admin/index-stats` | Chunks per file (`?domain_id=&tenant_id=`, `X-Admin-Secret`) |
| `POST /admin/reindex` | Rebuild vector index (`X-Admin-Secret`) |

Go calls: `PYTHON_RAG_URL` → `http://python:5000/rag/context`.

**No CV/classify in core** — RAG retrieval only. LLM stays in Go.

---

## `POST /rag/context`

JSON body:

```json
{
  "question": "...",
  "domain_id": "default",
  "tenant_id": "default",
  "locale": "en"
}
```

Response: `success`, `context`, `few_shot`, `fragments[]`, `category`, `error`.

Few-shot loaded from `config/locales/{locale}/few_shot.json`.

---

## gRPC Retriever

Service `grounded.rag.v1.Retriever` — see `api/proto/retriever.proto`.

```bash
grpcurl -plaintext -d '{"query":"How many vacation days?","domain_id":"default","tenant_id":"default","locale":"en","top_k":4}' \
  localhost:50051 grounded.rag.v1.Retriever/Retrieve
```

When `RAG_SERVICE_TOKEN` is set, pass metadata `x-rag-service-token`.

---

## `POST /admin/reindex`

Header `X-Admin-Secret` = env `ADMIN_SECRET`.

Chain: `reset_vector_store()` → `load_vector_store(force_reindex=True)`.

Indexes files from `data/{tenant_id}/{domain_id}/`: `.txt`, `.pdf`, `.docx`.

---

## Environment variables

| Variable | Purpose |
|----------|---------|
| `PYTHON_SERVICE_PORT` | HTTP port (default 5000) |
| `PYTHON_GRPC_PORT` | gRPC port (default 50051) |
| `REDIS_URL` | embedding cache (optional) |
| `EMBEDDING_CACHE_TTL_SEC` | embedding TTL (default 3600) |
| `DOMAINS_CONFIG_PATH` | path to `domains.json` |
| `LOCALES_ROOT` | path to locale bundles |
| `DEFAULT_LOCALE` | default few-shot locale (`en`) |
| `DEFAULT_TENANT_ID` | default tenant |
| `ADMIN_SECRET` | protect reindex |
| `RAG_SERVICE_TOKEN` | internal auth (Go ↔ Python / gRPC) |
| `FORCE_RAG_REINDEX` | full rebuild on startup |
| `GUNICORN_WORKERS` / `GUNICORN_TIMEOUT` | HTTP workers |

---

## Run

```bash
# from repo root — HTTP only (dev)
python -m flask --app api.app run -p 5000

# production-shaped (same as Docker): Gunicorn + gRPC
sh api/entrypoint.sh
```

Docker: `CMD ["/app/api/entrypoint.sh"]` in `Dockerfile.python`.

---

## What to read next

| Topic | File |
|-------|------|
| Providers / Redis / gRPC | [../LLM_PROVIDERS.md](../LLM_PROVIDERS.md) |
| Indexing | [rag-vector_store.md](./rag-vector_store.md) |
| Search | [rag-retrieval.md](./rag-retrieval.md) |
| Domains | [rag-domains_config.md](./rag-domains_config.md) |
| Docker | [docker-overview.md](./docker-overview.md) |
