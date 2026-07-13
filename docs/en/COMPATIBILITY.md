# Compatibility matrix

Supported stack for **Grounded LLM** reference implementation. Other versions may work but are not tested in CI.

Last updated: Phase 11 / main (2026-07)

---

## Platform runtimes

| Component | Supported | CI pin | Notes |
|-----------|-----------|--------|-------|
| **Go** | 1.23.x | `1.23` in `.github/workflows/ci.yml` | Server orchestration |
| **Python** | 3.11 – 3.12 | `3.11` in CI | RAG service only |
| **PostgreSQL** | 16.x | `postgres:16-alpine` | Sessions, messages, audit |
| **Node** | — | not required | Static webapp, no build step |

---

## ML / retrieval

| Component | Pin | Location |
|-----------|-----|----------|
| **Embedding model** | `intfloat/multilingual-e5-small` | `rag/vector_store.py` |
| **Vector store** | Chroma (default) or Qdrant (optional) | `VECTOR_STORE`, see [VECTOR_STORE.md](./docs/en/VECTOR_STORE.md) |
| **Chunking** | 500 / overlap 50 | `RecursiveCharacterTextSplitter` |
| **Hybrid rerank** | Keyword overlap (optional) | `RAG_RETRIEVAL_MODE=hybrid` or `RAG_RERANKER=keyword` |
| **Cross-encoder rerank** | Optional ML rerank | `RAG_RERANKER=cross_encoder` |

Changing the embedding model requires **full reindex** and eval gate re-run. Document the change in CHANGELOG and bump compatibility table.

---

## Container images (reference)

| Image | Base | Dockerfile |
|-------|------|------------|
| Server | `alpine:3.21` | `Dockerfile.server` |
| Python RAG | project-specific | `Dockerfile.python` |
| Webapp | nginx alpine | `Dockerfile.webapp` |

Release tags `v*.*.*` publish to GHCR (see `.github/workflows/release.yml`).

---

## LLM providers (operator choice)

Any **OpenAI-compatible** HTTPS endpoint:

| Variable | Example |
|----------|---------|
| `LLM_BASE_URL` | `https://openrouter.ai/api` |
| `LLM_MODEL` | operator-selected |

Not pinned — verify numeric grounding via built-in verify layer regardless of provider.

---

## CI / smoke modes

| Mode | Env | Use |
|------|-----|-----|
| Mock LLM + RAG | `LLM_MOCK=true`, `RAG_MOCK=true` | Unit tests, smoke, conformance |
| Retrieval eval | Python :5000 + Chroma | `eval-retrieval-gate` job |
| LLM E2E nightly | real `LLM_API_KEY` | optional secret |

---

## Operating systems

| OS | Support |
|----|---------|
| Linux (amd64) | Primary — Docker, K8s, CI |
| macOS | Dev (Docker Desktop) |
| Windows | Dev (Docker Desktop; native Go/Python for tests) |

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
curl -sS http://localhost:5000/health
go version   # expect 1.23+
python --version  # expect 3.11+
```

Report compatibility gaps via GitHub issues.
