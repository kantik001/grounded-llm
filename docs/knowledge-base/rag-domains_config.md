# Разбор: `rag/domains_config.py` / Domain catalog

**Исходный файл / Source:** `rag/domains_config.py`  
**Конфиг / Config:** `config/domains.json`  
**Кто использует / Used by:** `rag/vector_store.py`, `rag/retrieval.py`, Go (`server/domains.go`), тесты

---

## Назначение / Purpose

Загрузка каталога **knowledge domains** (workspace / база знаний).

---

## API

| Функция | Описание |
|---------|----------|
| `load_domains_config()` | читает JSON, кэш по mtime |
| `normalize_domain_id(id)` | trim, lower, проверка существования |
| `get_domain(id)` | метаданные домена |
| `list_domains()` | `{ default_domain, domains }` |
| `default_domain_id()` | обычно `"default"` |

---

## Формат `domains.json`

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

---

## Флаги домена / Domain flags

| Поле | Эффект |
|------|--------|
| `rag_enabled: false` | retrieval вернёт ошибку «база не подключена» |
| `rag_k` | число фрагментов retrieval (1–20, default 8) |
| `name` / `name_ru` | отображение в UI |

---

## Env

| Переменная | Путь по умолчанию |
|------------|-------------------|
| `DOMAINS_CONFIG_PATH` | `config/domains.json` |

В Docker: `/config/domains.json`.

---

## Go-зеркало / Go mirror

`server/domains.go` — тот же каталог для API `GET /domains` и guards в `domain_guards.go`.

После правки JSON: перезапуск Python; Go — SIGHUP или `CONFIG_RELOAD_INTERVAL_SEC`.
