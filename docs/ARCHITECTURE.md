# Архитектура: Grounded LLM / Architecture

Репозиторий — **platform core** для grounded-ассистентов в любой отрасли.  
Product packs (HR, legal, support и т.д.) — **domain pack**: `config/` + `data/{domain_id}/`.

---

## Слои / Layers

```
┌─────────────────────────────────────────────────────────┐
│  Platform core (this repo / этот репозиторий)           │
│  Go orchestration · Python RAG · verify · admin · CI      │
└───────────────────────────┬─────────────────────────────┘
                            │
              ┌─────────────┴─────────────┐
              ▼                           ▼
        Domain pack A              Domain pack B
        config + data/               config + data/
```

| Слой / Layer | Папки / Paths | Меняется? |
|--------------|---------------|-----------|
| **Core** | `server/`, `api/`, `rag/`, `migrations/`, `webapp/`, `scripts/` | Нет |
| **Domain pack** | `config/domains.json`, `prompts.json`, `branding.json`, `data/*` | **Да** |
| **Optional** | Vision/CV (вне ядра) | По задаче |

**`domain_id`** — идентификатор workspace / базы знаний.

---

## Поток (текст) / Text flow

1. Client → Go `POST /message`
2. Go → Python `POST /rag/context` (`domain_id`)
3. Chroma → фрагменты + few-shot
4. Go → LLM → verify → disclaimer → Postgres

---

## Документы KB / Knowledge documents

Форматы: **`.txt`**, **`.pdf`**, **`.docx`** → `rag/document_loaders.py` → chunking → Chroma.

---

## Чеклист нового домена / New domain checklist

1. Запись в `config/domains.json`
2. Документы в `data/{domain_id}/`
3. `prompts.json`, `few_shot.json`, `onboarding.json`, `branding.json`
4. `python scripts/reindex_rag.py`
5. `eval/rag_{domain}_baseline.jsonl` + `make eval-retrieval`

Оценка MVP: **2–5 дней** при готовых документах.

---

## Legacy

API и JSON: alias `crop_id`, `GET /crops` — совместимость со старыми клиентами.

---

## Документация / Docs

Язык: **RU + EN** (постепенный перевод). Подробности: `docs/knowledge-base/README.md`.
