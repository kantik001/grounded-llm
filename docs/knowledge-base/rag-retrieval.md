# Разбор: `rag/retrieval.py`

**Исходный файл:** `rag/retrieval.py`  
**Эндпоинт:** `POST /rag/context` в `api/app.py`  
**Дальше:** Go `server/rag_chat.go` собирает промпт и зовёт LLM

---

## Зачем этот файл

Слой **retrieval** в классической схеме RAG:

1. Принять вопрос и `crop_id`.
2. Найти фрагменты в Chroma (`vector_store.search`).
3. Собрать **context** для LLM.
4. Подобрать **few-shot** пример по типу вопроса.
5. Отдать JSON Go — **без генерации ответа**.

---

## Главная функция: `retrieve_rag_context(user_question, crop_id)`

### Вход

- `user_question` — текст от пользователя;
- `crop_id` — культура (по умолчанию `apple`).

### Выход (словарь)

| Поле | Назначение |
|------|------------|
| `success` | удалось ли собрать контекст |
| `error` | текст ошибки на русском |
| `context` | большой текст из фрагментов для промпта |
| `few_shot` | пример вопрос-ответ из `config/few_shot.json` |
| `category` | `fertilizer` / `disease` / `variety` / `general` |
| `fragments` | список `{filename, content}` для верификации на Go |
| `crop_id` | нормализованный id |

### Шаги внутри

1. Пустой вопрос → `success: false`.
2. `normalize_crop_id` — неверная культура → ошибка.
3. `get_crop` → если `rag_enabled: false` → «база статей не подключена».
4. `search(q, crop_id, k=8)` — векторный поиск.
5. Нет фрагментов → «Не нашёл информации в статьях…».
6. Склейка контекста:

```
Текст из статьи 'article1.txt':
<чанк>

---

Текст из статьи 'article2.txt':
...
```

7. `classify_question(q)` → категория.
8. `few_shot_for(crop_id, category)` → строка-пример для промпта.

---

## Классификация вопроса: `classify_question`

**Rule-based** (ключевые слова), не ML:

| Категория | Примеры слов в вопросе |
|-----------|-------------------------|
| `fertilizer` | удобрение, доза, азот, подкормк… |
| `disease` | болезн, парша, пятна, гниль, лечени… |
| `variety` | сорт, рентабельность, склон, Триумф… |
| `general` | всё остальное |

Зачем: в `few_shot.json` разные примеры тона и детализации по теме.

---

## Few-shot: `few_shot_for`

Читает `config/few_shot.json`:

```json
{
  "apple": {
    "fertilizer": "Пример вопроса: ... Пример ответа: ...",
    "disease": "...",
    "general": "..."
  }
}
```

Берёт категорию; если нет — fallback на `general`.

Кэш `_few_shot_cache` — один раз за процесс.

---

## Связь с Go

```mermaid
sequenceDiagram
    participant Go as server/rag_chat.go
    participant Py as retrieval.py
    participant Chroma as vector_store

    Go->>Py: POST /rag/context {question, crop_id}
    Py->>Chroma: search
    Chroma-->>Py: fragments
    Py-->>Go: context, few_shot, fragments
    Go->>Go: LLM + verifyRAGAnswer + disclaimer
```

Go **не** ходит в Chroma напрямую — только через Python.

---

## Логи

```
[RAG:apple] источник: article1.txt
```

Помогает отладке: какие чанки попали в контекст (в чат пользователю имена статей не обязательны).

---

## Ошибки vs «нет в материалах»

| Ситуация | Где решается |
|----------|----------------|
| Нет чанков в Chroma | `error` здесь, Go не зовёт LLM с пустым контекстом |
| LLM выдумал цифру | `verifyRAGAnswer` в Go (+ дубль логики в `verifier.py` для тестов) |
| Нет фактов в статьях, но LLM ответил | промпт в `rag_chat.go`: «нет в справочных материалах» |

---

## Что читать дальше

| Тема | Файл |
|------|------|
| Chroma, chunking | [rag-vector_store.md](./rag-vector_store.md) |
| Верификация чисел | [rag-verifier.md](./rag-verifier.md), `server/rag_chat.go` |
| HTTP | [python-api.md](./python-api.md) |

---

## Краткий итог

`retrieval.py` — **оркестратор RAG-поиска**: культура → поиск → context + few-shot + fragments для Go. Генерация ответа — не здесь.
