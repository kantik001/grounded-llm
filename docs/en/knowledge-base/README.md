# Project knowledge base (English)

Documentation for the **Grounded LLM** platform core.

**See also:** [../ARCHITECTURE.md](../ARCHITECTURE.md), [../DEPLOY.md](../DEPLOY.md), [../../eval/README.md](../../eval/README.md).  
Russian docs: [../../ru/knowledge-base/README.md](../../ru/knowledge-base/README.md).

---

## Contents

### Map and infrastructure

| Document | Description |
|----------|-------------|
| [PROJECT_STRUCTURE.md](./PROJECT_STRUCTURE.md) | Repository map |
| [docker-overview.md](./docker-overview.md) | Docker Compose services |
| [github-ci.yml.md](./github-ci.yml.md) | GitHub Actions CI |
| [config-overview.md](./config-overview.md) | `config/` and locales |
| [data-pipeline.md](./data-pipeline.md) | KB documents → RAG |
| [migrations-overview.md](./migrations-overview.md) | SQL migrations |

### Python RAG

| Document | Description |
|----------|-------------|
| [python-api.md](./python-api.md) | `api/app.py` |
| [rag-domains_config.md](./rag-domains_config.md) | `domains.json`, tenants |
| [rag-vector_store.md](./rag-vector_store.md) | Chroma, reindex |
| [rag-retrieval.md](./rag-retrieval.md) | `POST /rag/context` |
| [rag-verifier.md](./rag-verifier.md) | Answer verification |

### Go backend

| Document | Description |
|----------|-------------|
| [server-overview.md](./server-overview.md) | `server/*.go` overview |
| [server-auth-and-limits.md](./server-auth-and-limits.md) | Auth, API keys, CORS |
| [server-chat-and-db.md](./server-chat-and-db.md) | Sessions, Postgres, citations |
| [server-rag_chat.md](./server-rag_chat.md) | RAG + LLM + streaming |
| [server-admin-and-ux-api.md](./server-admin-and-ux-api.md) | Admin, metrics, OpenAPI |

### UI, scripts, quality

| Document | Description |
|----------|-------------|
| [webapp-overview.md](./webapp-overview.md) | Telegram Web App |
| [scripts-overview.md](./scripts-overview.md) | reindex, eval, init_domain |
| [tests-overview.md](./tests-overview.md) | pytest + Go tests |
| [quality-eval-and-rag-logs.md](./quality-eval-and-rag-logs.md) | Eval baselines |

Vision/CV is **not** in the core; use a domain pack when needed.

---

## Article naming

`{module}-{file}.md` maps to source files, e.g. `server-rag_chat.md` → `server/rag_chat.go`.
