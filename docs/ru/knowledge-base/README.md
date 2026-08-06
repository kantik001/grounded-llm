# База знаний по проекту

Документация для изучения и сопровождения **ядра платформы Grounded LLM** (на русском).

**См. также:** [../README.md](../README.md) (путь чтения), [../ARCHITECTURE.md](../ARCHITECTURE.md), [../DEPLOY.md](../DEPLOY.md), [../GUARDRAILS.md](../GUARDRAILS.md), [../../eval/README.md](../../eval/README.md).  
English: [../../en/knowledge-base/README.md](../../en/knowledge-base/README.md).

---

## Рекомендуемый порядок

1. [PROJECT_STRUCTURE.md](./PROJECT_STRUCTURE.md) — что где лежит  
2. [docker-overview.md](./docker-overview.md) — сервисы Compose  
3. [data-pipeline.md](./data-pipeline.md) — документ → чанк → ответ  
4. Python: [rag-domains_config](./rag-domains_config.md) → [rag-vector_store](./rag-vector_store.md) → [rag-retrieval](./rag-retrieval.md) → [rag-verifier](./rag-verifier.md)  
5. Go: [server-overview](./server-overview.md) → [server-chat-and-db](./server-chat-and-db.md) → [server-rag_chat](./server-rag_chat.md) → [server-auth-and-limits](./server-auth-and-limits.md)  
6. Опционально: [server-oidc-saas-analytics](./server-oidc-saas-analytics.md), [quality-eval-and-rag-logs](./quality-eval-and-rag-logs.md)

---

## Содержание

### Карта и инфраструктура

| Документ | Описание |
|----------|----------|
| [PROJECT_STRUCTURE.md](./PROJECT_STRUCTURE.md) | Карта репозитория (`packs/`, `connectors/`, `deploy/`, …) |
| [docker-overview.md](./docker-overview.md) | Docker Compose (+ опц. guardrails `:50052`) |
| [github-ci.yml.md](./github-ci.yml.md) | GitHub Actions CI (Go 1.25, 99 eval) |
| [config-overview.md](./config-overview.md) | `config/` и локали |
| [data-pipeline.md](./data-pipeline.md) | Документы KB → RAG → чат |
| [migrations-overview.md](./migrations-overview.md) | SQL-миграции `001`–`011` |

### Python RAG

| Документ | Описание |
|----------|----------|
| [python-api.md](./python-api.md) | HTTP `:5000` + gRPC Retriever `:50051` (`api/`) |
| [rag-domains_config.md](./rag-domains_config.md) | `domains.json`, tenant |
| [rag-vector_store.md](./rag-vector_store.md) | Chroma / Qdrant / pgvector, hybrid, reindex |
| [rag-retrieval.md](./rag-retrieval.md) | `POST /rag/context` |
| [rag-verifier.md](./rag-verifier.md) | Числа + faithfulness + NLI + remote `:50052` |

Ops-гайд по env: [../VECTOR_STORE.md](../VECTOR_STORE.md).

### Go backend (`server/internal/*`)

| Документ | Описание |
|----------|----------|
| [server-overview.md](./server-overview.md) | Пакеты `cmd/server` + `internal/` |
| [server-auth-and-limits.md](./server-auth-and-limits.md) | Telegram, API keys, CORS, rate limit |
| [server-chat-and-db.md](./server-chat-and-db.md) | Сессии, Postgres, citations, streaming |
| [server-rag_chat.md](./server-rag_chat.md) | RAG + LLM + verify |
| [server-admin-and-ux-api.md](./server-admin-and-ux-api.md) | Админка, metrics, onboarding |
| [server-oidc-saas-analytics.md](./server-oidc-saas-analytics.md) | OIDC, SaaS/Stripe, analytics, audit, metrics |

Связанные гайды: [../SAAS.md](../SAAS.md) · [../BILLING.md](../BILLING.md) · [../ANALYTICS_GUIDE.md](../ANALYTICS_GUIDE.md) · [../GUARDRAILS.md](../GUARDRAILS.md)

### UI, скрипты, качество

| Документ | Описание |
|----------|----------|
| [webapp-overview.md](./webapp-overview.md) | Чат, админка, signup, embed |
| [scripts-overview.md](./scripts-overview.md) | reindex, eval, init_pack, connectors |
| [tests-overview.md](./tests-overview.md) | pytest + Go tests |
| [quality-eval-and-rag-logs.md](./quality-eval-and-rag-logs.md) | **99** retrieval cases + логи `[RAG]` |

---

## Норматив (EN)

| Тема | EN |
|------|-----|
| Spec v1, conformance | [GROUNDED_SPEC_v1.md](../../en/spec/GROUNDED_SPEC_v1.md) |
| RFC-0001 | [RFC-0001](../../en/rfcs/RFC-0001-grounded-compatible.md) |
| Trust / governance | [TRUST_CENTER](../../en/TRUST_CENTER.md) · [GOVERNANCE](../../en/GOVERNANCE.md) |

Vision/CV **не входит в ядро** — отдельный domain pack.

---

## Именование статей

`{area}-{topic}.md` — модуль или поток. Пример: `server-rag_chat.md` → `internal/rag/pipeline.go` + `internal/httpapi/*`.  
Канон layout: [`server/README.md`](../../../server/README.md).
