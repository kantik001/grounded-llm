# Развёртывание / Deployment

Инструкция для **нового проекта** на каркасе Grounded LLM.  
Архитектура: [ARCHITECTURE.md](./ARCHITECTURE.md).

---

## Быстрый старт (Docker)

```bash
cp .env.example .env
# LLM_API_KEY, TELEGRAM_BOT_TOKEN (или TELEGRAM_AUTH_DISABLED=true для dev)

docker compose up -d --build
```

| Сервис | URL |
|--------|-----|
| Web App | http://localhost/ |
| Go API | http://localhost:8080/health |
| Python RAG | http://localhost:5000/health |

После добавления документов в `data/`:

```bash
python scripts/reindex_rag.py
# или POST /admin/reindex (Basic auth + ADMIN_SECRET для Python)
```

Поддерживаемые форматы KB: **`.txt`**, **`.pdf`**, **`.docx`**.

---

## Конфиг без пересборки

Каталог `./config` смонтирован в контейнеры как `/config` (read-only).

| Переменная | Файл |
|------------|------|
| `DOMAINS_CONFIG_PATH` | `domains.json` |
| `PROMPTS_CONFIG_PATH` | `prompts.json` |
| `ONBOARDING_CONFIG_PATH` | `onboarding.json` |
| `BRANDING_CONFIG_PATH` | `branding.json` |

Legacy alias: `CROPS_CONFIG_PATH` → тот же `domains.json`.

**Перезагрузка Go без рестарта:**

```bash
docker compose kill -s HUP server
```

Или `CONFIG_RELOAD_INTERVAL_SEC=300` в `.env`.

Python `rag/domains_config.py` перечитывает `domains.json` при изменении mtime.

---

## Локальная разработка (без Docker)

1. Postgres + `.env` с `DATABASE_URL`.
2. `cd server && go run .`
3. Python: `python api/app.py` (из корня репозитория).
4. Web: nginx или `webapp/` + `TELEGRAM_AUTH_DISABLED=true`, API на `:8080`.

---

## Eval после изменений KB

```bash
pip install requests
set PYTHON_RAG_URL=http://localhost:5000/rag/context
python scripts/run_rag_eval.py --suite default
make eval-retrieval
```

Результаты: `eval/results/YYYYMMDD_HHMMSS.json`.

Запускать после: reindex, смены `prompts.json`, смены `LLM_MODEL`.

---

## Новый заказчик: domain pack

### 1. Репозиторий

```bash
git clone <url> client-assistant
cd client-assistant
```

### 2. Domain pack

| Действие | Путь |
|----------|------|
| Документы KB | `data/{domain_id}/` (`.txt`, `.pdf`, `.docx`) |
| Каталог доменов | `config/domains.json` |
| Промпты и few-shot | `config/prompts.json`, `few_shot.json` |
| UI бренд | `config/branding.json`, при необходимости `webapp/` |
| Eval-вопросы | `eval/rag_{domain}_baseline.jsonl` |

### 3. Индексация и проверка

```bash
python scripts/reindex_rag.py
python scripts/run_rag_eval.py --suite default
```

### 4. Секреты

`.env`: `LLM_API_KEY`, `DATABASE_URL`, `CORS`, Telegram, `ADMIN_PASSWORD`, `ADMIN_SECRET`.

### 5. Пилот

Метрики: verify pass rate, доля «нет в материалах», 👍/👎, latency p95.

---

## Smoke

```bash
make smoke
# TELEGRAM_AUTH_DISABLED=true, server на :8080
```

---

## Что не переносить на новый инстанс

- volume `chroma_data` (пересоздаётся reindex).
- `postgres_data` / прод-сессии.
- Секреты `.env` — только шаблон `.env.example`.

---

## Опциональные модули

**Vision / CV** — не входит в ядро. Подключается отдельным domain pack (см. архивные статьи `cv-*.md` в knowledge-base).
