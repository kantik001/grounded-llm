# Архитектура Grounded LLM

Репозиторий — **ядро платформы** для ассистентов с ответами по документам (RAG) в любой отрасли.  
Отраслевые пакеты (HR, юриспруденция, поддержка и т.д.) — это **domain pack**: `config/` + `data/{tenant_id}/{domain_id}/`.

---

## Слои

```
┌─────────────────────────────────────────────────────────┐
│  Ядро платформы (этот репозиторий)                      │
│  Go · Python RAG · verify · админка · CI                │
└───────────────────────────┬─────────────────────────────┘
                            │
              ┌─────────────┴─────────────┐
              ▼                           ▼
        Domain pack A              Domain pack B
        config + data/               config + data/
```

| Слой | Папки | Меняется часто? |
|------|-------|-----------------|
| **Ядро** | `server/`, `api/`, `rag/`, `migrations/`, `webapp/`, `scripts/` | Нет |
| **Domain pack** | `config/domains.json`, `config/locales/{ru,en}/`, `data/*` | **Да** |
| **Опционально** | Vision/CV (вне ядра) | По задаче |

- **`domain_id`** — идентификатор домена / базы знаний.
- **`tenant_id`** — изоляция арендаторов (мультитенантность, фаза 2).

---

## Поток текстового чата

1. Клиент → Go `POST /message` (опционально `?stream=1` для потокового ответа)
2. Go → Python `POST /rag/context` (`domain_id`, `tenant_id`, `locale`)
3. Chroma → фрагменты документов + few-shot для выбранного языка
4. Go → LLM → проверка чисел → дисклеймер → Postgres (с `citations[]`)

Язык ответа и UI задаётся локалью (`ru` / `en`): см. `config/locales/`.

---

## Документы базы знаний

Форматы: **`.txt`**, **`.pdf`**, **`.docx`** → `rag/document_loaders.py` → разбиение на фрагменты → Chroma.

Рекомендуемый путь: `data/{tenant_id}/{domain_id}/`.  
Старый layout `data/{domain_id}/` по-прежнему поддерживается.

---

## Новый ассистент из template pack

Предпочтительно [packs/](../../packs/):

```bash
python scripts/init_pack.py list
python scripts/init_pack.py install it_support   # или: hr, legal_faq
python scripts/reindex_rag.py
```

Реестр: `packs/registry.yaml` — `python scripts/init_pack.py registry --validate`

---

## Чеклист нового домена (вручную)

1. Запись в `config/domains.json` (поля `names.ru` / `names.en` для UI)
2. Документы в `data/{tenant_id}/{domain_id}/`
3. Файлы в `config/locales/ru/` и `config/locales/en/` (prompts, few_shot, onboarding, branding)
4. `python scripts/reindex_rag.py` или `scripts/init_domain.ps1`
5. `eval/rag_{domain}_baseline.jsonl` + `make eval-retrieval`

Оценка MVP: **2–5 дней** при готовых документах.

---

## Документация

| Документация | [knowledge-base/README.md](./knowledge-base/README.md) |
| Указатель RU | [README.md](./README.md) |
| English | [../en/knowledge-base/README.md](../en/knowledge-base/README.md) |

Общий указатель: [../README.md](../README.md).
