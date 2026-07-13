# База знаний по проекту

Документация для изучения и сопровождения **ядра платформы Grounded LLM** (на русском языке).

**См. также:** [../README.md](../README.md), [../ARCHITECTURE.md](../ARCHITECTURE.md), [../DEPLOY.md](../DEPLOY.md), [../../eval/README.md](../../eval/README.md).  
English: [../../en/knowledge-base/README.md](../../en/knowledge-base/README.md).

### Только на английском (канон)

| Тема | EN |
|------|-----|
| Spec v1, conformance | [GROUNDED_SPEC_v1.md](../../en/spec/GROUNDED_SPEC_v1.md) |
| Terraform, K8s | [TERRAFORM.md](../../en/TERRAFORM.md), [K8S_DEPLOY.md](../../en/K8S_DEPLOY.md) |
| Trust center, backup | [TRUST_CENTER.md](../../en/TRUST_CENTER.md) |
| SDK | [QUICKSTART_SDK.md](../../en/QUICKSTART_SDK.md) |

---

## Содержание

### Карта и инфраструктура

| Документ | Описание |
|----------|----------|
| [PROJECT_STRUCTURE.md](./PROJECT_STRUCTURE.md) | Карта репозитория |
| [docker-overview.md](./docker-overview.md) | Docker Compose, 4 сервиса |
| [github-ci.yml.md](./github-ci.yml.md) | GitHub Actions CI |
| [config-overview.md](./config-overview.md) | `config/` и локали |
| [data-pipeline.md](./data-pipeline.md) | Документы KB → RAG → чат |
| [migrations-overview.md](./migrations-overview.md) | SQL-миграции PostgreSQL |

### Python RAG

| Документ | Описание |
|----------|----------|
| [python-api.md](./python-api.md) | HTTP-сервис `api/app.py` |
| [rag-domains_config.md](./rag-domains_config.md) | `domains.json`, tenant |
| [rag-vector_store.md](./rag-vector_store.md) | Chroma, индексация |
| [rag-retrieval.md](./rag-retrieval.md) | `POST /rag/context` |
| [rag-verifier.md](./rag-verifier.md) | Проверка чисел в ответе |

**Рекомендуемый порядок чтения:** domains_config → vector_store → retrieval → verifier → `server/rag_pipeline.go`

### Go backend

| Документ | Описание |
|----------|----------|
| [server-overview.md](./server-overview.md) | Обзор `server/*.go` |
| [server-auth-and-limits.md](./server-auth-and-limits.md) | Auth, API keys, CORS, лимиты |
| [server-chat-and-db.md](./server-chat-and-db.md) | Сессии, Postgres, citations, streaming |
| [server-rag_chat.md](./server-rag_chat.md) | RAG + LLM + verify |
| [server-admin-and-ux-api.md](./server-admin-and-ux-api.md) | Админка, metrics, onboarding |

### UI, скрипты, качество

| Документ | Описание |
|----------|----------|
| [webapp-overview.md](./webapp-overview.md) | Чат, админка, signup, embed |
| [scripts-overview.md](./scripts-overview.md) | reindex, eval, init_pack, connectors |
| [tests-overview.md](./tests-overview.md) | pytest + Go tests |
| [quality-eval-and-rag-logs.md](./quality-eval-and-rag-logs.md) | eval, логи `[RAG]` |

---

Vision/CV **не входит в ядро** — подключается отдельным domain pack.

---

## Именование статей

`{module}-{file}.md` соответствует исходнику, напр. `server-rag_chat.md` → код в `server/rag_pipeline.go` и `server/rag_chat.go`.
