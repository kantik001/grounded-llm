# Структура проекта Grounded LLM

Карта репозитория. Подробные разборы: [README.md](./README.md).

---

## Корень

| Путь | Назначение |
|------|------------|
| `server/` | Go: auth, сессии, RAG+LLM, админка, verify — [`server/README.md`](../../../server/README.md) (`cmd/server` + `internal/`) |
| `api/` | Python RAG **service** (HTTP+gRPC); см. `api/README.md` |
| `proto/` | Guardrails gRPC IDL; Retriever в `api/proto/` — см. `proto/README.md` |
| `rag/` | Retrieval **engine** (библиотека); HTTP/gRPC в `api/` — см. [`rag/README.md`](../../../rag/README.md) |
| `config/` | Domain pack + `locales/{ru,en}/` |
| `data/{tenant}/{domain}/` | Demo-файлы в git (SoT = Postgres + blobs) |
| `webapp/` | Эталонный UI — [`webapp/README.md`](../../../webapp/README.md) |
| `site/` | GitHub Pages лендинг — [`site/README.md`](../../../site/README.md) |
| `migrations/` | Схема PostgreSQL — см. [`migrations/README.md`](../../../migrations/README.md) |
| `eval/`, `scripts/`, `tests/` | Качество и эксплуатация — README в каждой папке |
| `packs/` | Официальные template packs + `init_pack.py` — [`packs/README.md`](../../../packs/README.md) |
| `connectors/` | Ingest (SharePoint, Drive, Confluence) — [`connectors/README.md`](../../../connectors/README.md) |
| `conformance/` | Spec / OpenAPI CLI — [`conformance/README.md`](../../../conformance/README.md) |
| `deploy/` | Helm + Terraform — [`deploy/README.md`](../../../deploy/README.md) |
| `sdk/python/`, `sdk/js/` | Клиенты Go API — [`sdk/README.md`](../../../sdk/README.md) |
| `models/`, `sparse_index/` | Runtime-бинарники/кэши (в gitignore) — см. README папок |
| `docs/ru/`, `docs/en/` | Документация на двух языках |

---

## `server/` — Go backend

Модуль **`grounded_llm_server`**: `cmd/server` + `internal/` (см. [`server/README.md`](../../../server/README.md)).

| Путь | Роль |
|------|------|
| `cmd/server` | точка входа |
| `internal/app` | composition: `Run()`, `Deps`, bridges |
| `internal/{config,store,auth,guardrails,metrics,llm,rag,httpapi,locale,domain,tenant,admin,oidc,saas,audit,analytics}` | доменные пакеты |
| `gen/guardrails/v1` | guardrails gRPC stubs |

→ [server-overview.md](./server-overview.md) · [server-oidc-saas-analytics.md](./server-oidc-saas-analytics.md)

---

## `rag/` — RAG-движок

Библиотека для `api/` и скриптов — не сетевой сервис. Карта + env: [`rag/README.md`](../../../rag/README.md).

| Модуль | Роль |
|--------|------|
| `vector_backend/` | Chroma / Qdrant / pgvector |
| `vector_store.py` | поиск, hybrid RRF, readiness |
| `indexing.py` / `document_loaders.py` | загрузка + чанки KB |
| `retrieval.py` | контекст + few-shot |
| `sparse_index.py` / `rerank.py` | BM25 hybrid + rerank |
| `verifier.py` | локальная проверка чисел |

---

## `config/` — domain pack

`domains.json`, `locales/ru/`, `locales/en/`, `examples/`, `schemas/` — см. [config/README.md](../../../config/README.md)

→ [config-overview.md](./config-overview.md)

---

## Документация

| Путь | Содержание |
|------|------------|
| `docs/ru/ARCHITECTURE.md` | Ядро vs domain pack |
| `docs/ru/DEPLOY.md` | Развёртывание |
| `docs/ru/knowledge-base/` | Разбор модулей |
| `docs/en/` | English mirror |

---

## Вне ядра

Computer Vision и отраслевые domain packs — **отдельные репозитории/пакеты**, не часть platform core.
