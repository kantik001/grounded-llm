# Разбор: папка `config/` / Config directory

**Папка / Folder:** `config/` — JSON без пересборки (volume `/config` в Docker).  
**Кто читает / Readers:** Go (`server/`), Python (`rag/`)

---

## Файлы в ядре / Core files

| Файл | Кто | Назначение |
|------|-----|------------|
| `domains.json` | Go + Python | Каталог доменов, `rag_enabled` |
| `prompts.json` | Go | System prompts, RAG constraints |
| `few_shot.json` | Python | Примеры для промпта LLM |
| `onboarding.json` | Go | Стартовые вопросы в Web App |
| `branding.json` | Go | Заголовки, дисклеймер UI |

---

## `domains.json`

```json
{
  "default_domain": "default",
  "domains": {
    "default": {
      "name": "Knowledge base",
      "emoji": "📚",
      "rag_enabled": true
    }
  }
}
```

- **Go:** `GET /domains`, guards `requireRAGEnabled`
- **Python:** `normalize_domain_id`, фильтр Chroma по `domain_id`

Env: `DOMAINS_CONFIG_PATH` (Docker: `/config/domains.json`).

Подробнее: [rag-domains_config.md](./rag-domains_config.md).

---

## `prompts.json`

Структура:

- `_platform.rag_constraints` — общие правила (не выдумывать, русский язык)
- `{domain_id}.rag_system`, `rag_task_intro` — промпты домена

Go: `loadPromptCatalog()`, `promptsForDomain()`.

---

## `few_shot.json`

```json
{
  "default": {
    "general": "Пример вопроса и ответа..."
  }
}
```

Python: `few_shot_for(domain_id)` в `rag/retrieval.py`.

---

## `onboarding.json`

Map `domain_id` → массив строк-вопросов.

`GET /onboarding?domain_id=default`

---

## `branding.json`

UI: `app_title`, `header_subtitle`, `disclaimer`, `onboarding_title`, …

`GET /branding`

---

## Перезагрузка / Reload

| Что изменили | Действие |
|--------------|----------|
| domains, prompts, onboarding, branding | Go: SIGHUP или `CONFIG_RELOAD_INTERVAL_SEC` |
| domains.json (Python) | restart `python` или mtime reload в `domains_config.py` |
| few_shot.json | restart `python` (кэш в retrieval) |
| документы в `data/` | **reindex** Chroma |

---

## Не в минимальном ядре

Отраслевые расширения (vision, отраслевые шаблоны) — **domain pack**, не файлы platform core.

---

## Новый домен / New domain checklist

1. Запись в `domains.json`
2. Блоки в `prompts.json`, `few_shot.json`, `onboarding.json`, `branding.json`
3. Документы в `data/{domain_id}/`
4. `python scripts/reindex_rag.py`
