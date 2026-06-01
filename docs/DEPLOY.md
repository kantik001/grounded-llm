# Развёртывание и клонирование платформы

Инструкция для **агробота** и для **нового проекта** на том же каркасе.  
Архитектура слоёв: [ARCHITECTURE.md](./ARCHITECTURE.md).

---

## Быстрый старт (Docker)

```bash
cp .env.example .env
# Заполнить LLM_API_KEY, TELEGRAM_BOT_TOKEN (или TELEGRAM_AUTH_DISABLED=true для dev)

docker compose up -d --build
```

| Сервис | URL |
|--------|-----|
| Web App | http://localhost/ |
| Go API | http://localhost:8080/health |
| Python | http://localhost:5000/health |

После добавления статей в `data/`:

```bash
python scripts/reindex_rag.py
# или POST /admin/reindex с Basic auth
```

---

## Конфиг без пересборки

Каталог `./config` смонтирован в контейнеры как `/config` (read-only).

| Переменная | Файл |
|------------|------|
| `CROPS_CONFIG_PATH` | `crops.json` |
| `PROMPTS_CONFIG_PATH` | `prompts.json` |
| `PHOTO_TEMPLATES_PATH` | `photo_templates.json` |
| `ONBOARDING_CONFIG_PATH` | `onboarding.json` |
| `BRANDING_CONFIG_PATH` | `branding.json` |

**Перезагрузка Go без рестарта:**

```bash
docker compose kill -s HUP server
```

Или `CONFIG_RELOAD_INTERVAL_SEC=300` в `.env`.

Python `rag/crops_config.py` перечитывает `crops.json` при изменении mtime.

---

## Локальная разработка (без Docker)

1. Postgres + `.env` с `DATABASE_URL`.
2. `cd server && go run .`
3. Python: `cd api` или корень с `FLASK_APP=api.app` / образ classifier.
4. Web: nginx или `webapp` + `TELEGRAM_AUTH_DISABLED=true`, API на `:8080`.

---

## Eval после изменений KB

```bash
# Только retrieval (Python :5000 должен быть доступен)
pip install requests
set CLASSIFIER_RAG_URL=http://localhost:5000/rag/context
python scripts/run_rag_eval.py --suite apple
python scripts/run_rag_eval.py --suite demo_hr

make eval-retrieval
```

Результаты: `eval/results/YYYYMMDD_HHMMSS.json`.

Гонять после: reindex, смены `prompts.json`, смены `LLM_MODEL`.

---

## Новый заказчик: клон платформы

### 1. Репозиторий

```bash
git clone <url> client-name-assistant
cd client-name-assistant
```

### 2. Domain pack

| Действие | Путь |
|----------|------|
| Удалить или заменить статьи | `data/` |
| Новые домены | `config/crops.json` |
| Промпты и few-shot | `config/prompts.json`, `few_shot.json` |
| UI бренд | `config/branding.json`, при необходимости `webapp/` |
| Выключить CV | `"cv_enabled": false` |
| Eval-вопросы | `eval/rag_<client>_baseline.jsonl` |

### 3. Индексация и проверка

```bash
python scripts/reindex_rag.py
python scripts/run_rag_eval.py --suite <client>
```

### 4. Секреты и регион

- `.env`: `LLM_API_KEY`, `DATABASE_URL`, CORS, Telegram или другой канал.
- Для KSA/GCC: хостинг в регионе (Bahrain/UAE), LLM в том же регионе, PDPL — отдельный договор.

### 5. Пилот

- Метрики: verify pass rate, доля «нет в материалах», 👍/👎, latency.
- Не логировать тело LLM (политика 1C в ROADMAP).

---

## Smoke

```bash
make smoke
# TELEGRAM_AUTH_DISABLED=true, server на :8080
```

---

## Что не копировать в новый проект

- `chroma_db/` volume (создаётся заново).
- `postgres_data` / прод-сессии.
- Секреты `.env` — только шаблон `.env.example`.
