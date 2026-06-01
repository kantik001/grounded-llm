# Структура проекта Grounded LLM / Repository map

Карта репозитория. Подробные разборы: [README.md](./README.md).

---

## Корень

| Путь | Назначение |
|------|------------|
| `server/` | Go: auth, sessions, RAG+LLM, admin, verify |
| `api/` | Python Flask: RAG retrieval |
| `rag/` | Chroma, retrieval, domains, document loaders |
| `config/` | Domain pack defaults |
| `data/{domain_id}/` | KB: `.txt`, `.pdf`, `.docx` |
| `webapp/` | Reference Telegram Web App UI |
| `migrations/` | PostgreSQL schema |
| `eval/`, `scripts/` | Quality & ops |
| `docs/` | Architecture, deploy, knowledge-base |

---

## `server/` — Go backend

Один бинарник `package main`, модуль **`grounded_llm_server`**.

Ключевые файлы:

| Файл | Роль |
|------|------|
| `main.go`, `routes.go` | старт, маршруты |
| `domains.go`, `domain_resolve.go` | каталог доменов, `domain_id` в API |
| `config_paths.go` | поиск JSON в `/config` |
| `rag_chat.go`, `rag_verify.go` | RAG + LLM + verify |
| `postgres_store.go` | Postgres, сессии, сообщения |
| `admin.go` | upload KB, reindex |

→ [server-overview.md](./server-overview.md)

---

## `rag/` — RAG engine

| Модуль | Роль |
|--------|------|
| `domains_config.py` | `config/domains.json` |
| `document_loaders.py` | `.txt`, `.pdf`, `.docx` |
| `vector_store.py` | Chroma indexing |
| `retrieval.py` | context для Go |
| `verifier.py` | verify чисел (mirror Go) |

---

## `config/` — domain pack

`domains.json`, `prompts.json`, `few_shot.json`, `onboarding.json`, `branding.json`

→ [config-overview.md](./config-overview.md)

---

## Документация

| Путь | Содержание |
|------|------------|
| `docs/ARCHITECTURE.md` | core vs domain pack |
| `docs/DEPLOY.md` | развёртывание |
| `docs/knowledge-base/` | разбор модулей |

---

## Вне ядра

Computer Vision, agro-конфиги и отраслевые domain packs — **отдельные репозитории/пакеты**, не часть platform core.
