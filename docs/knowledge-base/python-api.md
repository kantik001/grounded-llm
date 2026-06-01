# Разбор: `api/app.py` / Python RAG API

**Исходный файл / Source:** `api/app.py`  
**Язык / Language:** Python (Flask)  
**Связанные модули / Related:** `rag/retrieval.py`, `rag/vector_store.py`, `rag/domains_config.py`, `rag/document_loaders.py`  
**Кто вызывает / Called by:** Go server (`server/rag_chat.go`, `server/admin.go`)

---

## Зачем этот файл / Purpose

Отдельный **Python HTTP-сервис** (порт **5000**, контейнер compose: **`python`**).

| Эндпоинт | Назначение |
|----------|------------|
| `POST /rag/context` | Retrieval: фрагменты статей для вопроса (без LLM) |
| `GET /domains` | Каталог доменов из `config/domains.json` |
| `GET /health` | Healthcheck |
| `POST /admin/reindex` | Пересборка Chroma (`X-Admin-Secret`) |

Go вызывает: `PYTHON_RAG_URL` → `http://python:5000/rag/context`.

**CV / classify** в ядре **нет** — только RAG retrieval.

---

## `POST /rag/context`

Тело JSON:

```json
{ "question": "...", "domain_id": "default" }
```

Ответ: `success`, `context`, `few_shot`, `fragments[]`, `category`, `error`.

---

## `POST /admin/reindex`

Заголовок `X-Admin-Secret` = env `ADMIN_SECRET`.

Цепочка: `reset_vector_store()` → `load_vector_store(force_reindex=True)`.

Индексирует файлы из `data/{domain_id}/`: `.txt`, `.pdf`, `.docx`.

---

## Переменные окружения / Env

| Переменная | Назначение |
|------------|------------|
| `PYTHON_SERVICE_PORT` | порт (default 5000) |
| `DOMAINS_CONFIG_PATH` | путь к `domains.json` |
| `ADMIN_SECRET` | защита reindex |
| `FORCE_RAG_REINDEX` | полная пересборка при старте |

---

## Запуск / Run

```bash
# из корня репозитория
python api/app.py
```

Docker: `CMD ["python", "api/app.py"]` в `Dockerfile.python`.

---

## Что читать дальше

| Тема | Файл |
|------|------|
| Индексация | [rag-vector_store.md](./rag-vector_store.md) |
| Поиск | [rag-retrieval.md](./rag-retrieval.md) |
| Домены | [rag-domains_config.md](./rag-domains_config.md) |
| Docker | [docker-overview.md](./docker-overview.md) |
