# Разбор: RAG и LLM на Go (`server/rag_chat.go`)

**Файл:** `server/rag_chat.go`  
**Python:** [rag-retrieval.md](./rag-retrieval.md), [rag-verifier.md](./rag-verifier.md)  
**Вызывается из:** `handleChat` (устарел), `handleTextMessage` (`message_handlers.go`)

---

## Роль файла

Go **не** ищет в Chroma сам. Цепочка:

1. **`fetchRAGContext`** → Python `POST /rag/context` → context, few_shot, fragments.
2. Сборка **промпта** (`buildRAGUserPrompt` + `config/prompts.json`).
3. **`callLLMCompletion`** → OpenRouter / OpenAI-compatible.
4. **Постобработка** — `cleanRAGAnswer`, `appendRAGDisclaimer`.
5. **`verifyRAGAnswer`** — числа только из fragments (как [rag/verifier.py](../rag/verifier.py)).

---

## `fetchRAGContext(question, cropID)`

```json
POST CLASSIFIER_RAG_URL
{ "question": "...", "crop_id": "apple" }
```

Ответ `pythonRAGContextResponse`:

| Поле | Использование |
|------|----------------|
| `success` / `error` | ошибка «нет статей» и т.д. |
| `context` | блок `<context>` в промпте |
| `few_shot` | блок `<examples>` |
| `fragments` | верификация чисел |

Таймаут HTTP: **120s**.

---

## Промпт LLM

### Шаблон `ragUserPromptTpl`

Секции:

- `<system>` — intro из `prompts.json` (`RAGTaskIntro`)
- `<context>` — фрагменты статей
- `<examples>` — few-shot
- `<constraints>` — не выдумывать, русский, без «вероятно», **без названий статей**
- Вопрос пользователя

Системное сообщение отдельно: `prompts.RAGSystem` из `promptsForCrop(cropID)`.

### История диалога

`answerWithRAG(q, cropID, history)`:

- `history` — предыдущие user/assistant из БД (`HistoryForLLM`).
- В `/chat` history = `nil`.
- В `/message` — передаётся prior для контекста мультитурного чата.

---

## Постобработка ответа

### `cleanRAGAnswer`

Удаляет мусор модели: ``, `<answer>`, вступления «Давайте посмотрим…», «аботчик».

### `appendRAGDisclaimer`

1. `stripSourceAttribution` — строки `Источник: ...`
2. Добавляет константу:

> Справочная информация из базы знаний. Не заменяет очный осмотр агронома…

Пользователь **не** видит названия статей (политика продукта).

---

## `verifyRAGAnswer`

1. Собрать текст всех `fragments[].content`.
2. Извлечь числа из ответа (без дисклеймера) — regex, запятая → точка.
3. Каждое число в ответе должно совпасть с числом в контексте (±0.01).

**Провал:** пользователю сообщение вида «⚠️ Система не смогла подтвердить ответ…» + причина (число 72 не найдено).

**Без чисел в ответе** — verify OK.

Тесты: `rag_chat_test.go` (зеркало Python `test_verifier.py`).

---

## `answerWithRAG` — итоговая функция

Возвращает `(answer, success, errMsg, ragSoftFail)`:

| Случай | Поведение |
|--------|-----------|
| Пустой вопрос | error |
| RAG `success: false` | `ragSoftFail=true`, текст из Python |
| Нет `LLM_API_KEY` | error про ключ |
| LLM ошибка | error |
| Verify fail | answer с ⚠️, `success=true` (мягкий отказ) |
| OK | answer с дисклеймером |

---

## `handleChat` — `POST /chat` (устарел)

Ответ содержит заголовки `Deprecation: true` и `Link: </message>; rel="successor-version"`. Для новых интеграций используйте `POST /message` с `session_id`.

JSON: `{ "question", "crop_id" }`.

- Без сохранения в Postgres.
- Удобно для отладки RAG+LLM.
- Те же коды ответа: 400, 503 (нет ключа), 200 + `answer`.

---

## Связь с `handleTextMessage`

Тот же `answerWithRAG`, но:

- до/после — `AppendMessage` в БД;
- analytics `rag_answer`;
- ответ клиенту — весь массив `messages`.

---

## Требования для работы

| Компонент | Env / сервис |
|-----------|----------------|
| RAG retrieval | classifier up, статьи + reindex |
| LLM | `LLM_API_KEY` |
| Промпты | `PROMPTS_CONFIG_PATH` или `config/prompts.json` |
| Культура | `normalizeCropID` + `rag_enabled` |

---

## Отладка

1. Лог `RAG fetch error` — Python/Chroma.
2. Ответ «нет в справочных материалах» — RAG пустой или LLM по промпту.
3. ⚠️ verify — число в ответе не из статей (классический кейс «72%»).
4. 503 — нет `LLM_API_KEY`.

---

## Краткий итог

`rag_chat.go` — **мозг текстового ассистента на Go**: Python даёт факты, LLM формулирует, Go чистит и проверяет числа. Дублирует идеи `rag/verifier.py`, но в проде работает именно Go-версия.
