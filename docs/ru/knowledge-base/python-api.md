# `api/app.py` — Python RAG API

**Исходник:** `api/app.py`  
**Связанные модули:** `rag/retrieval.py`, `rag/vector_store.py`, `rag/domains_config.py`  
**Вызывает:** Go (`server/rag_chat.go`, `server/admin.go`)

---

## Назначение

Отдельный **Python HTTP-сервис** (порт **5000**, сервис Compose **`python`**).

| Эндпоинт | Назначение |
|----------|------------|
| `POST /rag/context` | Retrieval: фрагменты для вопроса (без LLM) |
| `GET /domains` | Каталог из `domains.json` |
| `GET /health` | Healthcheck |
| `GET /admin/index-stats` | Chunks по файлам (`?domain_id=&tenant_id=`, `X-Admin-Secret`) |
| `POST /admin/reindex` | Пересборка Chroma |

Go: `PYTHON_RAG_URL` → `http://python:5000/rag/context`.

**CV / classify в ядре нет** — только RAG retrieval.

---

## `POST /rag/context`

Тело JSON:

```json
{
  "question": "...",
  "domain_id": "default",
  "tenant_id": "default",
  "locale": "ru"
}
```

Ответ: `success`, `context`, `few_shot`, `fragments[]`, `category`, `error`.

Few-shot из `config/locales/{locale}/few_shot.json`.

---

## `POST /admin/reindex`

Заголовок `X-Admin-Secret` = env `ADMIN_SECRET`.

Индексирует `data/{tenant_id}/{domain_id}/`: `.txt`, `.pdf`, `.docx`.

---

## Переменные окружения

| Переменная | Назначение |
|------------|------------|
| `PYTHON_SERVICE_PORT` | Порт (5000) |
| `DOMAINS_CONFIG_PATH` | `domains.json` |
| `LOCALES_ROOT` | Путь к локалям |
| `DEFAULT_LOCALE` | Локаль few-shot по умолчанию |
| `DEFAULT_TENANT_ID` | Tenant по умолчанию |
| `ADMIN_SECRET` | Защита reindex |
| `FORCE_RAG_REINDEX` | Полная пересборка при старте |

---

## Запуск

```bash
# из корня репозитория
python api/app.py
```

Docker: `CMD ["python", "api/app.py"]` в `Dockerfile.python`.

---

## Дальше

| Тема | Файл |
|------|------|
| Индексация | [rag-vector_store.md](./rag-vector_store.md) |
| Поиск | [rag-retrieval.md](./rag-retrieval.md) |
| Docker | [docker-overview.md](./docker-overview.md) |
