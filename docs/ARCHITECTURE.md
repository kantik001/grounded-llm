# Архитектура: Grounded LLM (универсальное ядро)

Репозиторий — **platform core** для grounded-ассистентов в любой отрасли.  
Продуктовые пакеты (агро, HR, legal и т.д.) — это **domain pack**: `config/` + `data/{domain_id}/`.

## Слои

```
┌─────────────────────────────────────────────────────────┐
│  Platform core (этот репозиторий)                        │
│  Go orchestration · Python RAG · verify · admin · CI    │
└───────────────────────────┬─────────────────────────────┘
                            │
              ┌─────────────┴─────────────┐
              ▼                           ▼
        Domain pack A              Domain pack B
        config + data/               config + data/
```

| Слой | Папки | Меняется при новом проекте? |
|------|-------|----------------------------|
| **Core** | `server/`, `api/`, `rag/`, `migrations/`, `webapp/`, `scripts/` | Нет |
| **Domain pack** | `config/domains.json`, `prompts.json`, `branding.json`, `data/*` | **Да** |
| **Optional** | Vision/CV-модуль (вынесен из ядра; подключается отдельно) | По задаче |

**`domain_id`** — идентификатор workspace / базы знаний.

## Поток (текст)

1. Клиент → Go `POST /message`
2. Go → Python `POST /rag/context` (`domain_id`)
3. Chroma → фрагменты + few-shot
4. Go → LLM → verify → disclaimer → Postgres

## Чеклист нового домена

1. Запись в `config/domains.json`
2. Документы в `data/{domain_id}/`
3. `prompts.json`, `few_shot.json`, `onboarding.json`, `branding.json`
4. `python scripts/reindex_rag.py`
5. Свой `eval/rag_{domain}_baseline.jsonl` + `make eval-retrieval`

Оценка MVP: **2–5 дней** при готовых документах.
