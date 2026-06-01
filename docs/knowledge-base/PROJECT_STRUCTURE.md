# Структура проекта `doctor_gardens_ai`

Ниже — карта проекта по текущему состоянию репозитория: где какой файл и за что отвечает.

> Подробные разборы отдельных файлов: см. [README.md](./README.md) в этой папке.

## Корень проекта

- `.env.example` — пример переменных окружения для локального запуска и прод-конфига.
- `.gitignore` — исключения для Git (артефакты, секреты, временные файлы).
- `.dockerignore` — исключения при сборке Docker-образов.
- `README.md` — общее описание системы, запуск, API и архитектурный обзор.
- `Makefile` — удобные команды для разработки (тесты/запуск/утилиты).
- `docker-compose.yml`, `Dockerfile.*` → [docker-overview.md](./docker-overview.md)

## `docs/knowledge-base/` (эта папка)

- `README.md` — оглавление базы знаний.
- `PROJECT_STRUCTURE.md` — этот файл (карта проекта).
- `python-api.md` — разбор Python-сервиса `api/app.py`.
- `cv-apple_classifier.md` — разбор CV-модели `cv/apple_classifier.py`.
- `cv-registry.md` — фабрика моделей `cv/registry.py`.
- `cv-train_classifier.md` — обучение `cv/train_classifier.py`.
- `github-ci.yml.md` — разбор `.github/workflows/ci.yml`.
- `migrations-overview.md` — все SQL-миграции `migrations/*.sql`.
- `rag-crops_config.md`, `rag-vector_store.md`, `rag-retrieval.md`, `rag-verifier.md` — модули `rag/`.
- `scripts-overview.md` — утилиты `scripts/`.
- `tests-overview.md` — pytest в `tests/`.
- `webapp-overview.md` — фронтенд `webapp/`.
- `server-overview.md`, … — backend Go.
- `config-overview.md`, `docker-overview.md`, `data-pipeline.md`, `quality-eval-and-rag-logs.md` — конфиг, Docker, данные, план качества.

## `.github/workflows`

- `ci.yml` — GitHub Actions CI: тесты и проверка сборки. → [github-ci.yml.md](./github-ci.yml.md)

## `api/` (Python: HTTP для Go)

- `app.py` — Flask: `/classify`, `/rag/context`, `/health`, `/admin/reindex`. → [python-api.md](./python-api.md)

## `cv/` (Python: Computer Vision)

- `apple_classifier.py` — MobileNetV2, inference. → [cv-apple_classifier.md](./cv-apple_classifier.md)
- `labels_config.py` — метки классов из `cv_class_labels.json`
- `registry.py` — фабрика и кэш моделей по `crop_id`. → [cv-registry.md](./cv-registry.md)
- `train_classifier.py` — offline-обучение `.pth`. → [cv-train_classifier.md](./cv-train_classifier.md)
- `requirements.txt` — зависимости Python-сервиса (CV + RAG + Flask).

## `config/` (доменные и prompt-конфиги)

→ [config-overview.md](./config-overview.md)

- `crops.json`, `prompts.json`, `photo_templates.json`, `cv_class_labels.json`, `few_shot.json`, `onboarding.json`, `article_titles.json`, `branding.json`

## `data/` (база знаний для RAG)

### `data/apple/` (16 файлов)

- `article1.txt` … `article3.txt` — исходные статьи.
- `article4_scab.txt` … `article15_organic_calendar.txt` — болезни, уход, почва, вредители.
- `article16_planting_pit.txt` — посадочная яма (см. [data-pipeline.md](./data-pipeline.md)).

### `data/demo_hr/` (sandbox платформы)

- `policy_*.txt` — демо HR-политики; домен `demo_hr` в `crops.json` (`ui_hidden`, RAG без CV).
- Eval: [eval/rag_demo_hr_baseline.jsonl](../../eval/rag_demo_hr_baseline.jsonl).

### `data/pear/`

- `README.txt` — заметка/заглушка по заполнению контента для груши.

### `data/plum/`

- `README.txt` — заметка/заглушка по заполнению контента для сливы.

## `docs/` (обучающие и плановые документы)

- `ROADMAP.md` — общий план развития проекта по фазам.
- `ARCHITECTURE.md` — **ядро платформы vs domain pack**, чеклист клонирования.
- `DEPLOY.md` — развёртывание, reindex, eval, новый заказчик.
- `knowledge-base/` — база знаний по коду (эта папка).
- `LEARNING_SESSION_1.md` — итоги и выводы сессии 1.
- `LEARNING_SESSION_2.md` — итоги и выводы сессии 2.
- `LEARNING_SESSION_3.md` — итоги и выводы сессии 3.
- `LEARNING_SESSION_5.md` — итоги и выводы сессии 5.
- `LEARNING_SESSION_6.md` — итоги и выводы сессии 6.

## `migrations/` (SQL-миграции БД)

→ Общий разбор: [migrations-overview.md](./migrations-overview.md)

- `001_init.sql` — базовая схема (пользователи, сессии, сообщения).
- `002_crop_id.sql` — расширение схемы под мультикультуру (`crop_id`).
- `003_feedback_analytics.sql` — таблицы feedback/аналитики.

## `rag/` (Python: retrieval и верификация)

- `__init__.py` — пакетный маркер (пустой/минимальный, в базе знаний не разбирается).
- `crops_config.py` → [rag-crops_config.md](./rag-crops_config.md)
- `vector_store.py` → [rag-vector_store.md](./rag-vector_store.md)
- `retrieval.py` → [rag-retrieval.md](./rag-retrieval.md)
- `verifier.py` → [rag-verifier.md](./rag-verifier.md)
- `__pycache__/` — автокэш Python (см. FAQ в чате / не коммитить).

## `scripts/` (утилиты)

→ [scripts-overview.md](./scripts-overview.md)

- `reindex_rag.py` — принудительная переиндексация RAG-базы.
- `run_rag_eval.py` — прогон eval-наборов (`eval/*.jsonl`) → `eval/results/`.
- `smoke.sh` — smoke-проверки API для Linux/macOS.
- `smoke.ps1` — smoke-проверки API для Windows PowerShell.

## `eval/` (регрессии RAG)

→ [eval/README.md](../../eval/README.md)

- `rag_apple_baseline.jsonl` — 30 вопросов по яблоне.
- `rag_demo_hr_baseline.jsonl` — 5 вопросов sandbox HR.
- `results/` — отчёты прогонов.

## `server/` (backend API)

→ Начните с [server-overview.md](./server-overview.md)

| Файл | Статья |
|------|--------|
| `main.go`, `config.go`, `health.go` | [server-overview.md](./server-overview.md) — старт, конфиг, health |
| `llm.go`, `classifier_client.go`, `classify_flow.go`, `photo_recommendations.go`, `photo_templates.go`, `classify_handler.go` | [server-overview.md](./server-overview.md) — LLM и CV по фото |
| `auth_telegram.go`, `middleware.go`, `ratelimit.go` | [server-auth-and-limits.md](./server-auth-and-limits.md) |
| `message_handlers.go`, `session_handlers.go`, `chat_session.go`, `postgres_store.go` | [server-chat-and-db.md](./server-chat-and-db.md) |
| `rag_verify.go`, `rag_log.go`, `branding.go`, `crop_guards.go`, `api_errors.go`, `routes.go`, `config_reload.go` | [server-overview.md](./server-overview.md) |
| `rag_chat.go` | [server-rag_chat.md](./server-rag_chat.md) |
| `admin.go`, `onboarding.go`, `feedback.go`, `analytics_store.go`, `crops.go` | [server-admin-and-ux-api.md](./server-admin-and-ux-api.md) |
| `go.mod`, `go.sum` | зависимости Go |
| `*_test.go` | [tests-overview.md](./tests-overview.md) |

## `tests/` (Python-тесты)

→ [tests-overview.md](./tests-overview.md)

- `conftest.py` — pytest: `PYTHONPATH` на корень проекта.
- `test_crops_config.py` — тесты `rag/crops_config.py`.
- `test_verifier.py` — тесты `rag/verifier.py`.
- `requirements-test.txt` — pytest + langchain-core (без Chroma/PyTorch).

## `webapp/` (клиентский интерфейс)

→ [webapp-overview.md](./webapp-overview.md)

- `index.html`, `app.css`, `app.js` — Telegram Web App: чат, фото, онбординг, feedback.
- `admin.html` — админка: upload `.txt`, reindex (Basic auth).
- `nginx.conf` — прокси `/api/` → Go, раздача HTML.

---

## Как лучше изучать код (рекомендуемый порядок)

1. `README.md` → быстрый контекст по архитектуре.
2. [`ARCHITECTURE.md`](../ARCHITECTURE.md) → платформа vs domain pack, клонирование.
3. `docker-compose.yml` → как связаны сервисы.
4. [server-overview.md](./server-overview.md) → маршруты и старт.
5. [rag-vector_store.md](./rag-vector_store.md) → [rag-retrieval.md](./rag-retrieval.md) → `server/rag_chat.go` → ядро RAG.
6. [python-api.md](./python-api.md) + [cv-apple_classifier.md](./cv-apple_classifier.md) → CV-ветка.
7. `migrations/*.sql` + `server/postgres_store.go` → БД и персистентность.
8. `tests/`, `eval/`, `server/*_test.go` → качество и регрессии.
