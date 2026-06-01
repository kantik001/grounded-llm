# Папка `config/`

**Папка:** `config/` — JSON-конфиги без пересборки образов (в Docker монтируется как `/config`).  
**Читают:** Go (`server/`), Python (`rag/`)

---

## Основные файлы

| Путь | Кто | Назначение |
|------|-----|------------|
| `domains.json` | Go + Python | Каталог доменов, `rag_enabled`, `names.ru` / `names.en` |
| `locales/ru/prompts.json` | Go | Промпты RAG и правила ответа (русский) |
| `locales/en/prompts.json` | Go | То же для английского |
| `locales/*/few_shot.json` | Python | Few-shot примеры (`locale` в `POST /rag/context`) |
| `locales/*/onboarding.json` | Go | Стартовые вопросы в Web App |
| `locales/*/branding.json` | Go | Заголовки и дисклеймер UI |

Подробнее: [config/locales/README.md](../../../config/locales/README.md).

---

## `domains.json`

```json
{
  "default_domain": "default",
  "domains": {
    "default": {
      "name": "Knowledge base",
      "names": { "ru": "База знаний", "en": "Knowledge base" },
      "emoji": "📚",
      "rag_enabled": true
    }
  }
}
```

- **Go:** `GET /domains` — имя домена по языку запроса
- **Python:** фильтр Chroma по `domain_id` и `tenant_id`

Переменная: `DOMAINS_CONFIG_PATH` (в Docker: `/config/domains.json`).

→ [rag-domains_config.md](./rag-domains_config.md)

---

## Локали (`config/locales/`)

В каждой папке `ru/` и `en/`:

- `_platform.rag_constraints` — общие правила (язык ответа, без выдумок)
- `{domain_id}.rag_system`, `rag_task_intro` — промпты домена

Сервер: `initLocaleConfig()`, middleware по заголовкам `X-Locale`, `Accept-Language` или query `?locale=`.

Переменные: `DEFAULT_LOCALE` (по умолчанию `ru`), `LOCALES_ROOT`.

**Важно:** для русскоязычных пользователей правьте в первую очередь `locales/ru/prompts.json` — оттуда берётся язык ответа LLM.

---

## `few_shot.json` (внутри локали)

Примеры «вопрос → ответ» для подсказки модели. Python: `few_shot_for(domain_id, locale=...)`.

---

## Onboarding и branding

- `GET /onboarding?domain_id=default&locale=ru`
- `GET /branding?locale=ru`

---

## Перезагрузка

| Изменили | Действие |
|----------|----------|
| `domains.json` | Go: SIGHUP; Python: restart или mtime reload |
| `locales/*` | Go: reload; Python: restart (кэш few_shot) |
| файлы в `data/{tenant}/{domain}/` | **reindex** Chroma |

---

## Чеклист нового домена

1. Запись в `domains.json` (+ `names` для RU/EN)
2. Блоки в `locales/ru/` и `locales/en/`
3. Документы в `data/{tenant_id}/{domain_id}/`
4. `python scripts/reindex_rag.py` или `scripts/init_domain.ps1`
