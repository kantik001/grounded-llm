# `rag/retrieval.py` — RAG retrieval

**Исходник:** `rag/retrieval.py`  
**Эндпоинт:** `POST /rag/context`  
**Дальше:** Go `server/rag_pipeline.go` → LLM

---

## Назначение

1. Принять вопрос, `domain_id`, `tenant_id`, `locale`
2. `vector_store.search()` — фрагменты из Chroma
3. Собрать `context` + few-shot для локали
4. Вернуть JSON для Go — **без LLM**

---

## `retrieve_rag_context(...)`

### Поля ответа

| Поле | Назначение |
|------|------------|
| `success` | Контекст найден |
| `error` | Текст ошибки для пользователя |
| `context` | Текст для промпта |
| `few_shot` | Из `config/locales/{locale}/few_shot.json` |
| `category` | Сейчас `"general"` |
| `fragments` | `{filename, content, excerpt, page?}` для verify и citations |
| `domain_id` | Нормализованный id |

### Мягкие ошибки (soft fail)

- пустой вопрос
- неизвестный `domain_id`
- `rag_enabled: false`
- нет фрагментов в Chroma

---

## Few-shot

`few_shot_for(domain_id, category, locale)` — ключ домена в locale-файле.

---

## Связанные файлы

| Тема | Файл |
|------|------|
| Поиск | [rag-vector_store.md](./rag-vector_store.md) |
| Домены | [rag-domains_config.md](./rag-domains_config.md) |
| Go | [server-rag_chat.md](./server-rag_chat.md) |
