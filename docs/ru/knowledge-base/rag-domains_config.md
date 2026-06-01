# `rag/domains_config.py` — каталог доменов

**Исходник:** `rag/domains_config.py`  
**Конфиг:** `config/domains.json`  
**Используют:** `rag/vector_store.py`, `rag/retrieval.py`, Go (`server/domains.go`), тесты

---

## Назначение

Загрузка каталога **доменов знаний** (workspace / база знаний).

---

## API модуля

| Функция | Описание |
|---------|----------|
| `load_domains_config()` | Читает JSON, кэш по mtime |
| `normalize_domain_id(id)` | trim, lower, проверка существования |
| `get_domain(id)` | Метаданные домена |
| `list_domains()` | `{ default_domain, domains }` |
| `default_domain_id()` | Обычно `"default"` |

---

## Формат `domains.json`

```json
{
  "default_domain": "default",
  "domains": {
    "default": {
      "name": "Knowledge base",
      "names": { "ru": "База знаний", "en": "Knowledge base" },
      "emoji": "📚",
      "rag_enabled": true,
      "rag_k": 8
    }
  }
}
```

---

## Поля домена

| Поле | Эффект |
|------|--------|
| `rag_enabled: false` | retrieval вернёт «база не подключена» |
| `rag_k` | Число фрагментов (1–20, по умолчанию 8) |
| `names` / `name` / `name_ru` | Отображение в UI (Go учитывает locale) |
| `ui_hidden` | Скрыть из `GET /domains` |

---

## Переменные окружения

| Переменная | Путь по умолчанию |
|------------|-------------------|
| `DOMAINS_CONFIG_PATH` | `config/domains.json` |

В Docker: `/config/domains.json`.

---

## Зеркало на Go

`server/domains.go` — тот же каталог для `GET /domains` и проверок в `domain_guards.go`.

После правки JSON: перезапуск Python; Go — SIGHUP или `CONFIG_RELOAD_INTERVAL_SEC`.
