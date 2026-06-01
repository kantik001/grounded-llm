# База знаний / Project knowledge base

Документация для изучения кода **Grounded LLM platform core**.  
Язык: русский + английские термины в заголовках.

**См. также:** [../ARCHITECTURE.md](../ARCHITECTURE.md), [../DEPLOY.md](../DEPLOY.md), [../../eval/README.md](../../eval/README.md).

---

## Содержание

### Карта и инфраструктура

| Документ | Описание |
|----------|----------|
| [PROJECT_STRUCTURE.md](./PROJECT_STRUCTURE.md) | Карта репозитория |
| [docker-overview.md](./docker-overview.md) | Docker Compose, 4 сервиса |
| [github-ci.yml.md](./github-ci.yml.md) | GitHub Actions CI |
| [config-overview.md](./config-overview.md) | `config/*.json` |
| [data-pipeline.md](./data-pipeline.md) | Документы KB → RAG |
| [migrations-overview.md](./migrations-overview.md) | SQL-миграции |

### Python RAG

| Документ | Описание |
|----------|----------|
| [python-api.md](./python-api.md) | `api/app.py` |
| [rag-domains_config.md](./rag-domains_config.md) | `domains.json`, `domain_id` |
| [rag-vector_store.md](./rag-vector_store.md) | Chroma, loaders, reindex |
| [rag-retrieval.md](./rag-retrieval.md) | `POST /rag/context` |
| [rag-verifier.md](./rag-verifier.md) | verify чисел, disclaimer |

**Порядок:** domains_config → vector_store → retrieval → verifier → `server/rag_chat.go`

### Go backend

| Документ | Описание |
|----------|----------|
| [server-overview.md](./server-overview.md) | Обзор `server/*.go` |
| [server-auth-and-limits.md](./server-auth-and-limits.md) | Telegram, CORS, rate limit |
| [server-chat-and-db.md](./server-chat-and-db.md) | Сессии, Postgres |
| [server-rag_chat.md](./server-rag_chat.md) | RAG + LLM + verify |
| [server-admin-and-ux-api.md](./server-admin-and-ux-api.md) | Admin, onboarding, feedback |

### UI, scripts, quality

| Документ | Описание |
|----------|----------|
| [webapp-overview.md](./webapp-overview.md) | `webapp/`, nginx |
| [scripts-overview.md](./scripts-overview.md) | reindex, smoke, eval |
| [tests-overview.md](./tests-overview.md) | pytest + Go tests |
| [quality-eval-and-rag-logs.md](./quality-eval-and-rag-logs.md) | eval, логи `[RAG]` |

---

Vision/CV — **не входит в ядро**; подключается отдельным domain pack при необходимости.

---

## Именование новых статей

`{module}-{file}.md` → исходник в репозитории, напр. `server-rag_chat.md` → `server/rag_chat.go`.
