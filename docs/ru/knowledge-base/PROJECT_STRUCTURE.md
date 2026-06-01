# Структура проекта Grounded LLM

Карта репозитория. Подробные разборы: [README.md](./README.md).

---

## Корень

| Путь | Назначение |
|------|------------|
| `server/` | Go: auth, сессии, RAG+LLM, админка, verify |
| `api/` | Python Flask: retrieval |
| `rag/` | Chroma, поиск, домены, загрузчики документов |
| `config/` | Domain pack + `locales/{ru,en}/` |
| `data/{tenant}/{domain}/` | База знаний: `.txt`, `.pdf`, `.docx` |
| `webapp/` | Эталонный UI Telegram Web App |
| `migrations/` | Схема PostgreSQL |
| `eval/`, `scripts/` | Качество и эксплуатация |
| `docs/ru/`, `docs/en/` | Документация на двух языках |

---

## `server/` — Go backend

Один бинарник `package main`, модуль **`grounded_llm_server`**.

| Файл | Роль |
|------|------|
| `main.go`, `routes.go` | Старт, маршруты |
| `domains.go`, `locale.go` | Домены и локали |
| `rag_pipeline.go`, `rag_verify.go` | RAG + LLM + verify |
| `sse.go`, `api_keys.go`, `tenant.go` | Streaming, API keys, tenant |
| `postgres_store.go` | Postgres |
| `admin.go` | Upload KB, reindex |

→ [server-overview.md](./server-overview.md)

---

## `rag/` — RAG-движок

| Модуль | Роль |
|--------|------|
| `domains_config.py` | `config/domains.json` |
| `document_loaders.py` | `.txt`, `.pdf`, `.docx` |
| `vector_store.py` | Индексация Chroma, фильтр tenant |
| `retrieval.py` | Контекст для Go |
| `verifier.py` | Проверка чисел (эталон для тестов) |

---

## `config/` — domain pack

`domains.json`, `locales/ru/`, `locales/en/` (prompts, few_shot, onboarding, branding)

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
