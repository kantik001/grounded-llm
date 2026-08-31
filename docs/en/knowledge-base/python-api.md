# `api/` — Python RAG service (HTTP + gRPC)

**Package layout:** see [`api/README.md`](../../../api/README.md).  
**Source:** `api/http/app.py`, `api/grpc/retriever.py`, `api/entrypoint.sh`, `api/auth.py`  
**Language:** Python (Flask app module; production HTTP via **Gunicorn**)  
**Related:** `rag/retrieval.py`, `rag/embedding_cache.py`, `rag/vector_store.py`, `rag/domains_config.py`  
**Called by:** Go server (`PYTHON_RAG_URL`); agents via gRPC `:50051`

> Folder name `api` is historical. This is **not** the public product HTTP API (`server/` `/api/v1`). It is the internal retrieval service.

---

## Purpose

Separate **Python service** (Compose service **`python`**):

| Surface | Port | Role |
|---------|------|------|
| HTTP (Gunicorn) | **5000** | Retrieval + admin reindex / ingest |
| gRPC Retriever | **50051** | Agent-friendly `Retrieve` RPC |

| Endpoint | Purpose |
|----------|---------|
| `POST /rag/context` | Retrieval: article fragments for a question (no LLM) |
| `GET /domains` | Domain catalog (`RAG_SERVICE_TOKEN` when set) |
| `GET /health` | Liveness |
| `GET /ready` | Readiness: data dir + index/backend smoke (`RAG_SERVICE_TOKEN` when set) |
| `GET /metrics` | Retrieve counters/histogram (by protocol+outcome) + embedding cache |
| `GET /admin/index-stats` | Chunks per file (`?domain_id=&tenant_id=`, `X-Admin-Secret`) |
| `POST /admin/reindex` | Rebuild vector index (`X-Admin-Secret`; `409` if already running) |
| `POST /admin/ingest/run` | Start ingest job by `job_id` (internal; Go triggers after `ingest_jobs` insert) |
| `GET /admin/ingest/status` | Ingest job + tasks (`?job_id=`) |
| `GET /admin/ingest/metrics` | In-process counters + DLQ depth |

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

Service `grounded.rag.v1.Retriever` — contract: `api/proto/retriever.proto`.  
Generated stubs: `api/gen/retriever_pb2.py`, `api/gen/retriever_pb2_grpc.py` (do not edit by hand).

```bash
# After changing the .proto (from repo root):
pip install grpcio-tools==1.69.0
python scripts/gen_retriever_grpc.py
python scripts/gen_retriever_grpc.py --check   # CI uses this
```

`grpcio` / `grpcio-health-checking` / `grpcio-tools` are pinned to **1.69.0** (see `api/requirements.txt` and `Dockerfile.python`).

```bash
grpcurl -plaintext -d '{"query":"How many vacation days?","domain_id":"default","tenant_id":"default","locale":"en","top_k":4}' \
  localhost:50051 grounded.rag.v1.Retriever/Retrieve
```

When `RAG_SERVICE_TOKEN` is set, pass metadata `x-rag-service-token` or `authorization: Bearer <token>`.
Optional metadata `x-request-id` is echoed into logs. Business errors use `success=false` in the response; unexpected failures return gRPC `INTERNAL`.
`top_k` is passed as an explicit retrieve argument (not via process env). Thread pool size: `GRPC_MAX_WORKERS` (default 4).
TLS is not terminated in-process — use private network / mesh.
HTTP accepts the same: `X-RAG-Service-Token` or `Authorization: Bearer …`.
In production, both `RAG_SERVICE_TOKEN` and `ADMIN_SECRET` must be set and ≥ 16 characters.

---

## `POST /admin/reindex`

Header `X-Admin-Secret` = env `ADMIN_SECRET`.

Incremental refresh via `refresh_vector_store()` (or full when `{"full": true}`).

Concurrent calls return **409** (`reindex already in progress`).

---

## Ingest admin endpoints

Called by Go after `POST /api/admin/ingest` (see [INGESTION.md](../INGESTION.md)).

| Endpoint | Body / query | Role |
|----------|----------------|------|
| `POST /admin/ingest/run` | `{"job_id": 1, "sync": false}` | Start pipeline for existing Postgres job |
| `GET /admin/ingest/status` | `?job_id=` | Job status + per-file tasks |
| `GET /admin/ingest/metrics` | — | Prometheus-style lines, `dlq_depth` |

`sync: true` processes parse → embed → finalize in-process (no Redis worker).

Indexes documents from Postgres registry + blob store (formats `.txt`, `.pdf`, `.docx`).

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
| `ADMIN_SECRET` | protect reindex / ingest admin routes |
| `DATABASE_URL` | ingest job state (Postgres) |
| `INGEST_ASYNC` | `1` = Redis queues; `0` = sync drain in API process |
| `INGEST_STAGING_DIR` | parsed chunk JSONL staging |
| `RAG_SERVICE_TOKEN` | internal auth (Go ↔ Python / gRPC) |
| `FORCE_RAG_REINDEX` | full rebuild on startup |
| `GUNICORN_WORKERS` / `GUNICORN_TIMEOUT` | HTTP workers |
| `GRPC_MAX_WORKERS` | gRPC thread pool (default 4) |

---

## Run

```bash
# from repo root — HTTP only (dev)
python -m flask --app api.http.app run -p 5000

# gRPC only
python -m api.grpc

# production-shaped (same as Docker): Gunicorn + gRPC supervisor
sh api/entrypoint.sh
```

Docker (`Dockerfile.python`): `ENTRYPOINT ["tini", "--"]` + `CMD ["/app/api/entrypoint.sh"]` (tini = PID 1, signal forwarding / zombie reaping).

Compatibility shims: `api.app` and `api.grpc_retriever` still import the new modules.

---

## What to read next

| Topic | File |
|-------|------|
| Providers / Redis / gRPC | [../LLM_PROVIDERS.md](../LLM_PROVIDERS.md) |
| Indexing | [rag-vector_store.md](./rag-vector_store.md) · [INGESTION.md](../INGESTION.md) |
| Search | [rag-retrieval.md](./rag-retrieval.md) |
| Domains | [rag-domains_config.md](./rag-domains_config.md) |
| Docker | [docker-overview.md](./docker-overview.md) |
