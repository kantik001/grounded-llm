# `rag/verifier.py`

**Исходник:** `rag/verifier.py`  
**Тесты:** `tests/test_verifier.py`  
**В проде (по умолчанию):** проверка в Go — `server/rag_verify.go` (`GUARDRAILS_MODE=local`)  
**Опционально:** gRPC → [grounded-guardrails](https://github.com/kantik001/grounded-guardrails) `:50052` при `GUARDRAILS_MODE=remote|hybrid` — см. [../en/GUARDRAILS.md](../en/GUARDRAILS.md)

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

## Python vs Go vs guardrails

| | Python | Go (local) | grounded-guardrails |
|--|--------|------------|---------------------|
| Прод | эталон для тестов | **default** после RAG-ответа | опционально `remote` / `hybrid` |
| Логика | та же идея Spec ±0.01 | `verifyRAGAnswer` | Rust reference + Go host parity |

При изменениях держите tolerance **±0.01** согласованным.

---

## Тесты

`pytest tests/test_verifier.py` — без Chroma и LLM.

---

## Дальше

| Тема | Файл |
|------|------|
| Промпт | [server-rag_chat.md](./server-rag_chat.md) |
| Фрагменты | [rag-retrieval.md](./rag-retrieval.md) |
| Remote verify | [../en/GUARDRAILS.md](../en/GUARDRAILS.md) |
