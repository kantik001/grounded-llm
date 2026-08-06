# `rag/verifier.py`

**Исходник:** `rag/verifier.py`  
**Тесты:** `tests/test_verifier.py`  
**В проде (по умолчанию):** проверка в Go — `internal/rag/verify.go` (`GUARDRAILS_MODE=local`). Опционально: `VERIFY_FAITHFULNESS`, `VERIFY_NLI`.  
**Опционально:** gRPC → [grounded-guardrails](https://github.com/kantik001/grounded-guardrails) `:50052` при `GUARDRAILS_MODE=remote|hybrid` — см. [../GUARDRAILS.md](../GUARDRAILS.md)

---

## Зачем нужен

Защита от **галлюцинаций с цифрами**: если LLM написала «72%» или «748,5 см», это число должно **встречаться в найденных фрагментах**.

Названия файлов-источников пользователю **не показываются** — вместо них общий дисклеймер (добавляет Go).

---

## Константа `RAG_ANSWER_DISCLAIMER`

Текст в конце ответа:

> Справочная информация из базы знаний. Не заменяет официальную консультацию ответственного специалиста.

В Python используется при `strip_source_attribution`, чтобы числа из дисклеймера не мешали проверке.

---

## `extract_numbers(text)`

- Заменяет `,` на `.` (десятичные по-русски).
- Regex: `\b\d+(?:\.\d+)?\b`.
- Возвращает список `float`.

---

## `verify_answer(question, answer, fragments)`

1. Склеить `page_content` фрагментов → `context_text`.
2. Очистить ответ от служебных строк.
3. Извлечь числа из ответа и контекста.
4. Каждое число в ответе должно быть в контексте с точностью **±0.01**.
5. Лишние числа → `(False, "Число(а) [...] не найдены в источниках.")`.

---

## Faithfulness (`VERIFY_FAITHFULNESS`)

Лексическая проверка: каждое содержательное предложение ответа должно опираться на фрагменты (stem-based, RU/EN).

| Env | Поведение |
|-----|-----------|
| `off` | выкл. |
| `warn` (default) | лог, verify всё равно pass |
| `enforce` | fail при неподдерживаемых предложениях |

Код: `internal/rag/verify_faithfulness.go`.

---

## Опциональный NLI (`VERIFY_NLI`)

| Env | Поведение |
|-----|-----------|
| `off` (default) | только числа + faithfulness |
| `assist` | лексика сначала; NLI подтверждает fail |
| `replace` | только NLI |

`VERIFY_NLI_URL` — HTTP endpoint. Код: `internal/rag/verify_nli.go`.

---

## Python vs Go vs guardrails

| | Python | Go (`internal/rag/*`) | grounded-guardrails |
|--|--------|----------------------|---------------------|
| Прод | эталон для тестов | **default** после RAG | опц. `remote` / `hybrid` |
| Числа | Spec ±0.01 | `VerifyRAGAnswer` | gRPC `VerifyText` |
| Faithfulness | — | `VERIFY_FAITHFULNESS` | — |
| NLI | — | `VERIFY_NLI` | — |

---

## Тесты

`pytest tests/test_verifier.py` — без Chroma и LLM.

---

## Дальше

| Тема | Файл |
|------|------|
| Промпт / pipeline | [server-rag_chat.md](./server-rag_chat.md) |
| Фрагменты | [rag-retrieval.md](./rag-retrieval.md) |
| Remote verify | [../GUARDRAILS.md](../GUARDRAILS.md) |
