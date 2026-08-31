# Project knowledge base (English)

Documentation for the **Grounded LLM** platform core.

**See also:** [../ARCHITECTURE.md](../ARCHITECTURE.md), [../DEPLOY.md](../DEPLOY.md), [../GUARDRAILS.md](../GUARDRAILS.md), [../../eval/README.md](../../eval/README.md).  
Russian docs: [../../ru/knowledge-base/README.md](../../ru/knowledge-base/README.md).

---

## Contents

### Map and infrastructure

| Document | Description |
|----------|-------------|
| [PROJECT_STRUCTURE.md](./PROJECT_STRUCTURE.md) | Repository map (modules + paths) |
| [docker-overview.md](./docker-overview.md) | Docker Compose services |
| [github-ci.yml.md](./github-ci.yml.md) | GitHub Actions CI |
| [config-overview.md](./config-overview.md) | `config/` and locales |
| [data-pipeline.md](./data-pipeline.md) | KB documents → RAG |
| [../INGESTION.md](../INGESTION.md) | Async ingest jobs (parse/embed/index) |
| [../KB_SOURCE_OF_TRUTH.md](../KB_SOURCE_OF_TRUTH.md) | Document registry, blobs, index runs |
| [migrations-overview.md](./migrations-overview.md) | SQL migrations `001`–`013` |

### Python RAG

| Document | Description |
|----------|-------------|
| [python-api.md](./python-api.md) | Python RAG service (`api/`) |
| [rag-domains_config.md](./rag-domains_config.md) | `domains.json`, tenants |
| [rag-vector_store.md](./rag-vector_store.md) | Vector backends, reindex |
| [rag-retrieval.md](./rag-retrieval.md) | `POST /rag/context` |
| [rag-verifier.md](./rag-verifier.md) | Numeric + faithfulness + optional NLI / guardrails |

### Go backend (`server/internal/*`)

| Document | Description |
|----------|-------------|
| [server-overview.md](./server-overview.md) | Package layout (`cmd/server` + `internal/`) |
| [server-auth-and-limits.md](./server-auth-and-limits.md) | Telegram, API keys, OIDC, CORS, rate limit |
| [server-chat-and-db.md](./server-chat-and-db.md) | Sessions, Postgres, citations |
| [server-rag_chat.md](./server-rag_chat.md) | RAG + LLM + streaming + verify |
| [server-admin-and-ux-api.md](./server-admin-and-ux-api.md) | Admin, metrics, OpenAPI, branding |

**Related (top-level EN, not KB):** [SAAS.md](../SAAS.md) · [BILLING.md](../BILLING.md) · [ANALYTICS_GUIDE.md](../ANALYTICS_GUIDE.md) · [VECTOR_STORE.md](../VECTOR_STORE.md) · RU deep dive [server-oidc-saas-analytics.md](../../ru/knowledge-base/server-oidc-saas-analytics.md)

### UI, scripts, quality

| Document | Description |
|----------|-------------|
| [webapp-overview.md](./webapp-overview.md) | Chat UI, admin, embed, signup |
| [scripts-overview.md](./scripts-overview.md) | reindex, eval, init_pack |
| [tests-overview.md](./tests-overview.md) | pytest + Go tests |
| [quality-eval-and-rag-logs.md](./quality-eval-and-rag-logs.md) | **99** retrieval cases + `[RAG]` logs |

Vision/CV is **not** in the core; use a domain pack when needed.

---

## Article naming

`{area}-{topic}.md` maps to a **module or flow**, not a single flat file. Example:

- `server-rag_chat.md` → `internal/rag/pipeline.go` + `internal/httpapi/{message,sse}.go`
- `rag-retrieval.md` → `rag/retrieval.py` + `api/http/app.py`

Canonical layout: [`server/README.md`](../../../server/README.md).
