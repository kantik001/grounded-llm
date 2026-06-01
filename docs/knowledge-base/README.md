# База знаний по проекту / Project knowledge base

Документация для изучения кода. Язык: **русский** (основной) + **английские** заголовки и термины там, где указано.

**Platform core vs domain pack:** [../ARCHITECTURE.md](../ARCHITECTURE.md), [../DEPLOY.md](../DEPLOY.md).

---

## Содержание / Index

### Карта и инфраструктура

| Документ | Описание |
|----------|----------|
| [PROJECT_STRUCTURE.md](./PROJECT_STRUCTURE.md) | Карта репозитория |
| [docker-overview.md](./docker-overview.md) | Docker Compose, 4 сервиса, volumes |
| [github-ci.yml.md](./github-ci.yml.md) | GitHub Actions CI |
| [config-overview.md](./config-overview.md) | `config/*.json` |
| [data-pipeline.md](./data-pipeline.md) | Документы KB → RAG |
| [migrations-overview.md](./migrations-overview.md) | SQL-миграции |

### Python RAG (`api/`, `rag/`)

| Документ | Описание |
|----------|----------|
| [python-api.md](./python-api.md) | `api/app.py` — Flask RAG API |
| [rag-domains_config.md](./rag-domains_config.md) | `domains.json`, `domain_id` |
| [rag-vector_store.md](./rag-vector_store.md) | Chroma, loaders, reindex |
| [rag-retrieval.md](./rag-retrieval.md) | `POST /rag/context` |
| [rag-verifier.md](./rag-verifier.md) | verify чисел, disclaimer |

**Порядок RAG:** domains_config → vector_store → retrieval → verifier → `server/rag_chat.go`

### Go backend (`server/`)

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

### Опционально / Optional (legacy CV)

> Модуль **`cv/`** не входит в platform core. Статьи сохранены как reference для domain pack с vision.

| Документ | Описание |
|----------|----------|
| [cv-apple_classifier.md](./cv-apple_classifier.md) | PyTorch classifier (legacy) |
| [cv-registry.md](./cv-registry.md) | Model registry (legacy) |
| [cv-train_classifier.md](./cv-train_classifier.md) | Training (legacy) |

### Устаревшие ссылки

| Старое | Новое |
|--------|-------|
| `rag-crops_config.md` | [rag-domains_config.md](./rag-domains_config.md) |
| `crop_id` | `domain_id` |
| сервис `classifier` | сервис `python` |

---

## Как пользоваться / How to use

1. Не знаете, где код → **PROJECT_STRUCTURE.md**
2. Конкретный файл → соответствующий `*.md` в этой папке
3. Новый domain pack → [../ARCHITECTURE.md](../ARCHITECTURE.md) checklist

---

## Именование новых статей / Naming

`{module}-{file}.md` → исходник в репозитории, напр. `server-rag_chat.md` → `server/rag_chat.go`.

В начале статьи: **исходный файл**, **связанные модули**, краткий EN subtitle при необходимости.
