# Compatibility matrix

Supported stack for **Grounded LLM** reference implementation. Other versions may work but are not tested in CI.

Last updated: **v0.3.0** / main (2026-07-27)

---

## Platform runtimes

| Component | Supported | CI pin | Notes |
|-----------|-----------|--------|-------|
| **Go** | 1.25.x | `1.25` in `.github/workflows/ci.yml` | Server orchestration (`server/go.mod`) |
| **Python** | 3.11 – 3.12 | `3.11` in CI | RAG HTTP (Gunicorn) + gRPC Retriever |
| **PostgreSQL** | 16.x | `pgvector/pgvector:pg16` | Sessions, messages, audit; optional pgvector embeddings |
| **Redis** | 7.x | `redis:7-alpine` in Compose | Embedding + LLM response caches |
| **Node** | — | not required | Static webapp, no build step |

---

## ML / retrieval

| Component | Pin | Location |
|-----------|-----|----------|
| **Embedding model** | `intfloat/multilingual-e5-small` | vector backends under `rag/vector_backend/` |
| **Vector store** | Chroma (default), Qdrant, or pgvector (optional) | `VECTOR_STORE`, see [VECTOR_STORE.md](./VECTOR_STORE.md) |
| **Embedding cache** | Redis key `embedding:{md5}:{model}` | `REDIS_URL`, `EMBEDDING_CACHE_TTL_SEC` |
| **Chunking** | 500 / overlap 50 | `RecursiveCharacterTextSplitter` |
| **Hybrid rerank** | BM25 + dense + RRF | `RAG_RETRIEVAL_MODE=hybrid` |
| **Keyword rerank** | Overlap on query tokens (optional) | `RAG_RERANKER=keyword` |
| **Cross-encoder rerank** | Optional ML rerank | `RAG_RERANKER=cross_encoder` |

Changing the embedding model requires **full reindex** and eval gate re-run. Document the change in CHANGELOG and bump compatibility table.

---

## Container images (reference)

| Image | Base | Dockerfile |
|-------|------|------------|
| Server | `alpine:3.21` | `Dockerfile.server` |
| Python RAG | `python:3.11-slim` + Gunicorn/gRPC entrypoint | `Dockerfile.python` |
| Webapp | nginx alpine | `Dockerfile.webapp` |

Release tags `v*.*.*` publish to GHCR (see `.github/workflows/release.yml`).

---

## LLM providers (operator choice)

Any **OpenAI-compatible** `/v1/chat/completions` endpoint. Switch with env only — see [LLM_PROVIDERS.md](./LLM_PROVIDERS.md).

| `LLM_PROVIDER` | Default base | Notes |
|----------------|--------------|-------|
| `openai` (default) | `https://openrouter.ai/api` | Needs `LLM_API_KEY` |
| `ollama` | `http://ollama:11434` | Compose `--profile ollama`; key auto `local` |
| `vllm` | `http://vllm:8000` | Compose `--profile vllm`; NVIDIA |

Override anytime with `LLM_BASE_URL` / `LLM_MODEL`. Numeric grounding still applies via the verify layer.

---

## CI / smoke modes

| Mode | Env | Use |
|------|-----|-----|
| Mock LLM + RAG | `LLM_MOCK=true`, `RAG_MOCK=true` | Unit tests, smoke, conformance |
| Retrieval eval | Python :5000 + vector store | `eval-retrieval-gate` job (**99** cases) |
| LLM E2E nightly | real `LLM_API_KEY` | optional secret |

---

## Operating systems

| OS | Support |
|----|---------|
| Linux (amd64) | Primary — Docker, K8s, CI |
| macOS | Dev (Docker Desktop) |
| Windows | Dev (Docker Desktop; prefer Ollama profile for local LLM; native Go/Python for tests) |

---

## API version ↔ product release

| API path | OpenAPI file | Introduced |
|----------|--------------|------------|
| `/api/v1/*` | `server/openapi.v1.json` | Phase 2 |
| `/api/v1/signup`, `/api/v1/plans` | (see [SAAS.md](./SAAS.md)) | Phase 10 — optional |
| `/api/v1/billing/stripe/*` | (see [BILLING.md](./BILLING.md)) | Phase 10–11 — optional |

See [API_DEPRECATION_POLICY.md](./API_DEPRECATION_POLICY.md) for stability rules.

---

## Checking your deploy

```bash
curl -sS http://localhost:8080/health
curl -sS http://localhost:8080/ready
curl -sS http://localhost:8080/metrics   # llm_tokens_*, TTFT, caches
curl -sS http://127.0.0.1:5000/health
go version   # expect 1.25.x
python --version  # expect 3.11+
```

Optional gRPC (with `grpcurl`):

```bash
grpcurl -plaintext localhost:50051 list
```

Report compatibility gaps via GitHub issues.
