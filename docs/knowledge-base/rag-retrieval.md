# Разбор: `rag/retrieval.py` / RAG retrieval

**Исходный файл / Source:** `rag/retrieval.py`  
**Эндпоинт / Endpoint:** `POST /rag/context`  
**Дальше / Next:** Go `server/rag_chat.go` → LLM

---

## Назначение / Purpose

1. Принять вопрос и `domain_id`
2. `vector_store.search()` — фрагменты из Chroma
3. Собрать `context` + `few_shot`
4. JSON для Go — **без LLM**

---

## `retrieve_rag_context(question, domain_id)`

### Выход

| Поле | Назначение |
|------|------------|
| `success` | контекст найден |
| `error` | сообщение на русском |
| `context` | текст для промпта |
| `few_shot` | из `config/few_shot.json` |
| `category` | сейчас `"general"` |
| `fragments` | `{filename, content}` для verify |
| `domain_id` | нормализованный id |

### Ошибки (soft fail)

- пустой вопрос
- неизвестный `domain_id`
- `rag_enabled: false`
- нет фрагментов в Chroma

---

## Few-shot

`few_shot_for(domain_id, category)` — ключ домена в `few_shot.json`.

---

## Связанные файлы

| Тема | Файл |
|------|------|
| Search | [rag-vector_store.md](./rag-vector_store.md) |
| Domains | [rag-domains_config.md](./rag-domains_config.md) |
| Go orchestration | [server-rag_chat.md](./server-rag_chat.md) |
