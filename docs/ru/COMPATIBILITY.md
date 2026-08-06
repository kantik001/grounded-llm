**Канон (EN):** [COMPATIBILITY.md](../en/COMPATIBILITY.md)

# Матрица совместимости

Поддерживаемый стек референса **Grounded LLM**. Другие версии могут работать, но в CI не проверяются.

Обновлено: **v0.4.0** / main (2026-08-06)

---

## Рантаймы платформы

| Компонент | Supported | CI pin | Заметки |
|-----------|-----------|--------|---------|
| **Go** | 1.25.x | `1.25` в `.github/workflows/ci.yml` | Оркестратор (`server/go.mod`, `cmd/server`) |
| **Python** | 3.11 – 3.12 | `3.11` | RAG HTTP (Gunicorn) + gRPC Retriever |
| **PostgreSQL** | 16.x | `pgvector/pgvector:pg16` | Сессии, audit; опц. pgvector embeddings |
| **Redis** | 7.x | `redis:7-alpine` | Embedding + LLM response caches |
| **Node** | — | не нужен | Static webapp |

Миграции включают **`010_saas_tenants.sql`**, **`011_admin_users_membership.sql`** (и предыдущие) — см. [migrations-overview.md](./knowledge-base/migrations-overview.md).

---

## ML / retrieval

| Компонент | Pin | Где |
|-----------|-----|-----|
| **Embedding model** | `intfloat/multilingual-e5-small` | `rag/vector_backend/` |
| **Vector store** | Chroma (default), Qdrant, pgvector | `VECTOR_STORE` — [VECTOR_STORE.md](./VECTOR_STORE.md) |
| **Embedding cache** | Redis `embedding:{md5}:{model}` | `REDIS_URL`, `EMBEDDING_CACHE_TTL_SEC` |
| **Chunking** | 500 / overlap 50 | `RecursiveCharacterTextSplitter` |
| **Hybrid** | BM25 + dense + RRF | `RAG_RETRIEVAL_MODE=hybrid` |
| **Rerank** | keyword / cross_encoder | `RAG_RERANKER` |

Смена embedding-модели → **полный reindex** + eval gate. Зафиксировать в CHANGELOG.

---

## Образы (референс)

| Image | Base | Dockerfile |
|-------|------|------------|
| Server | `alpine:3.21` | `Dockerfile.server` |
| Python RAG | `python:3.11-slim` | `Dockerfile.python` |
| Webapp | nginx alpine | `Dockerfile.webapp` |

Теги `v*.*.*` → GHCR (`.github/workflows/release.yml`).

---

## LLM-провайдеры

Любой OpenAI-compatible `/v1/chat/completions`. Env only — [LLM_PROVIDERS.md](./LLM_PROVIDERS.md).

| `LLM_PROVIDER` | Default base | Notes |
|----------------|--------------|-------|
| `openai` | `https://openrouter.ai/api` | Нужен `LLM_API_KEY` |
| `ollama` | `http://ollama:11434` | `--profile ollama` |
| `vllm` | `http://vllm:8000` | `--profile vllm`; NVIDIA |

Numeric grounding — через слой verify независимо от провайдера.

---

## CI / smoke

| Mode | Env | Use |
|------|-----|-----|
| Mock LLM + RAG | `LLM_MOCK=true`, `RAG_MOCK=true` | Unit, smoke, conformance |
| Retrieval eval | Python `:5000` + vector store | `eval-retrieval-gate` (**99** cases) |
| LLM E2E nightly | реальный `LLM_API_KEY` | optional secret |

---

## ОС

| OS | Support |
|----|---------|
| Linux (amd64) | Primary — Docker, K8s, CI |
| macOS | Dev (Docker Desktop) |
| Windows | Dev (Docker Desktop; для local LLM — profile Ollama) |

---

## API ↔ релиз

| Path | Introduced |
|------|------------|
| `/api/v1/*` | Phase 2 |
| `/api/v1/signup`, `/api/v1/plans` | Phase 10 — optional ([SAAS.md](./SAAS.md)) |
| `/api/v1/billing/stripe/*` | Phase 10–11 — optional ([BILLING.md](./BILLING.md)) |

Политика: [API_DEPRECATION_POLICY.md (EN)](../en/API_DEPRECATION_POLICY.md).

---

## Проверка деплоя

```bash
curl -sS http://localhost:8080/health
curl -sS http://localhost:8080/ready
curl -sS http://localhost:8080/metrics
curl -sS http://127.0.0.1:5000/health
go version          # 1.25.x
python --version    # 3.11+
grpcurl -plaintext localhost:50051 list   # опционально
```
