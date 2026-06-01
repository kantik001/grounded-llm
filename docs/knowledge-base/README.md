# База знаний по проекту

Документация для самостоятельного изучения кода: вы или коллега можете открыть нужный файл и быстро понять, что делает модуль.

**Платформа (ядро vs domain pack):** [../ARCHITECTURE.md](../ARCHITECTURE.md), [../DEPLOY.md](../DEPLOY.md), [../eval/README.md](../eval/README.md).

## Содержание

| Документ | Описание |
|----------|----------|
| [PROJECT_STRUCTURE.md](./PROJECT_STRUCTURE.md) | Карта репозитория: папки и файлы, за что отвечает каждый |
| [python-api.md](./python-api.md) | Подробный разбор `api/app.py` (Python Flask, не Go) |
| [cv-apple_classifier.md](./cv-apple_classifier.md) | PyTorch MobileNetV2: классы болезней, inference, веса `.pth` |
| [cv-registry.md](./cv-registry.md) | Фабрика и кэш моделей по `crop_id`, `MODEL_PATH`, `cv_enabled` |
| [cv-train_classifier.md](./cv-train_classifier.md) | Обучение модели, датасет, сохранение `apple_classifier.pth` |
| [github-ci.yml.md](./github-ci.yml.md) | GitHub Actions CI: зачем, три job, когда запускается (без DevOps-жаргона) |
| [migrations-overview.md](./migrations-overview.md) | SQL-миграции 001–003: синтаксис, связи таблиц, как накатываются при старте |

### RAG (`rag/`, без `__init__.py`)

| Документ | Описание |
|----------|----------|
| [rag-crops_config.md](./rag-crops_config.md) | `config/crops.json`, `crop_id`, `rag_enabled` / `cv_enabled` |
| [rag-vector_store.md](./rag-vector_store.md) | Chroma, chunking, embeddings, reindex, `data/` → `chroma_db/` |
| [rag-retrieval.md](./rag-retrieval.md) | Поиск, context, few-shot, `POST /rag/context` |
| [rag-verifier.md](./rag-verifier.md) | Проверка чисел в ответе, дисклеймер (дубль логики на Go) |

**Порядок чтения RAG:** `crops_config` → `vector_store` → `retrieval` → `verifier` → `server/rag_chat.go`

### Утилиты (`scripts/`)

| Документ | Описание |
|----------|----------|
| [scripts-overview.md](./scripts-overview.md) | `reindex_rag.py`, `smoke.ps1`, `smoke.sh` — когда запускать, что проверяют |

### Тесты

| Документ | Описание |
|----------|----------|
| [tests-overview.md](./tests-overview.md) | `tests/` (pytest), что покрыто, запуск, связь с Go-тестами и CI |

### Web UI (`webapp/`)

| Документ | Описание |
|----------|----------|
| [webapp-overview.md](./webapp-overview.md) | `index.html`, `admin.html`, `nginx.conf` — чат, админка, прокси |

### Go backend (`server/`)

| Документ | Описание |
|----------|----------|
| [server-overview.md](./server-overview.md) | Старт, конфиг, маршруты, разбиение `server/*.go` (LLM, CV, health) |
| [server-auth-and-limits.md](./server-auth-and-limits.md) | Telegram initData, CORS, rate limit |
| [server-chat-and-db.md](./server-chat-and-db.md) | `POST /message`, Postgres, фото, сессии |
| [server-rag_chat.md](./server-rag_chat.md) | RAG + LLM + verify + дисклеймер |
| [server-admin-and-ux-api.md](./server-admin-and-ux-api.md) | Админка, crops, onboarding, feedback |

**Порядок чтения server:** overview → auth → chat-and-db → rag_chat (после Python RAG) → admin-and-ux

### Инфраструктура, конфиг, данные, качество

| Документ | Описание |
|----------|----------|
| [config-overview.md](./config-overview.md) | `config/*.json`: crops, prompts, branding, `photo_templates`, few-shot, onboarding |
| [docker-overview.md](./docker-overview.md) | docker-compose, 4 сервиса, volumes, порты, `.env` |
| [data-pipeline.md](./data-pipeline.md) | Загрузка статей `.txt`, reindex, обучение `.pth` |
| [quality-eval-and-rag-logs.md](./quality-eval-and-rag-logs.md) | Eval-наборы, `run_rag_eval.py`, логи `[RAG]` |

## Как пользоваться

1. Не знаете, где искать код → **PROJECT_STRUCTURE.md**.
2. Разбираете конкретный файл → откройте соответствующий `*.md` в этой папке (по мере наполнения).
3. План развития продукта → [`../ROADMAP.md`](../ROADMAP.md).
4. Итоги учебных сессий → `../LEARNING_SESSION_*.md`.

## Добавление новых статей

Именование: `{путь-к-файлу через дефис}.md`, например:

- `server-rag_chat.md` → `server/rag_chat.go`
- `python-api.md` → `api/app.py`
- `cv-registry.md` → `cv/registry.py`
- `rag-retrieval.md` → `rag/retrieval.py`

В начале каждой статьи указывайте **исходный файл** в репозитории и **связанные модули**.
