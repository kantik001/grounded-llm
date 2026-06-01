# Структура проекта Grounded LLM / Repository map

Карта репозитория по текущему состоянию. Подробные разборы: [README.md](./README.md).

---

## Корень / Root

| Файл | Назначение |
|------|------------|
| `README.md` | Обзор, quick start, API |
| `PROJECT_STRUCTURE.md` | Краткая карта (корень) |
| `docker-compose.yml` | 4 сервиса: postgres, python, server, webapp |
| `Dockerfile.server`, `Dockerfile.python`, `Dockerfile.webapp` | образы |
| `Makefile` | up, test, smoke, reindex |
| `.env.example` | шаблон секретов |

---

## `server/` — Go backend

Оркестрация: auth, sessions, RAG+LLM, verify, admin. → [server-overview.md](./server-overview.md)

Тесты: `server/*_test.go`

---

## `api/` — Python Flask

`app.py`: `/rag/context`, `/health`, `/admin/reindex`, `/domains`. → [python-api.md](./python-api.md)

Зависимости: `api/requirements.txt`

---

## `rag/` — RAG engine

| Модуль | Роль |
|--------|------|
| `domains_config.py` | `config/domains.json` |
| `document_loaders.py` | `.txt`, `.pdf`, `.docx` |
| `vector_store.py` | Chroma, indexing |
| `retrieval.py` | context для Go |
| `verifier.py` | verify чисел (Python mirror) |

---

## `config/` — domain pack (частично core defaults)

| Файл | Назначение |
|------|------------|
| `domains.json` | каталог доменов, `rag_enabled` |
| `prompts.json` | system prompts, constraints |
| `few_shot.json` | примеры для RAG |
| `onboarding.json` | стартовые вопросы |
| `branding.json` | UI брендинг |

→ [config-overview.md](./config-overview.md)

---

## `data/{domain_id}/` — knowledge base

Документы: `.txt`, `.pdf`, `.docx`. Демо: `data/default/` (HR policies).

---

## `webapp/` — reference UI

`index.html`, `admin.html`, `app.js`, `nginx.conf`. → [webapp-overview.md](./webapp-overview.md)

---

## `migrations/` — PostgreSQL

`001`…`004` — users, sessions, messages, feedback, analytics, `domain_id`. → [migrations-overview.md](./migrations-overview.md)

---

## `tests/` — pytest

`test_verifier.py`, `test_domains_config.py`, `test_document_loaders.py`. → [tests-overview.md](./tests-overview.md)

---

## `scripts/`

`reindex_rag.py`, `run_rag_eval.py`, `smoke.ps1`, `smoke.sh`. → [scripts-overview.md](./scripts-overview.md)

---

## `eval/`

Baseline JSONL для регрессии retrieval. → [../eval/README.md](../eval/README.md)

---

## `docs/`

| Путь | Содержание |
|------|------------|
| `docs/ARCHITECTURE.md` | слои core / domain pack |
| `docs/DEPLOY.md` | развёртывание |
| `docs/knowledge-base/` | разбор модулей (эта папка) |

---

## Нет в ядре / Not in core

- `cv/` — vision module (optional domain pack)
- `photo_templates.json`, `cv_class_labels.json` — legacy agro configs

---

## Legacy naming

| Было | Стало |
|------|-------|
| `doctor_gardens_ai` | `grounded-llm` |
| `crop_id` | `domain_id` |
| `classifier` (compose) | `python` |
| `crops.json` | `domains.json` |
