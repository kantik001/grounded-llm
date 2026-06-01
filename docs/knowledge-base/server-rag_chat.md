# Разбор: RAG и LLM / `server/rag_chat.go`

**Файл / Source:** `server/rag_chat.go`  
**Python:** [rag-retrieval.md](./rag-retrieval.md), [rag-verifier.md](./rag-verifier.md)  
**Вызывается из / Called by:** `handleTextMessage` (`message_handlers.go`)

---

## Цепочка / Pipeline

1. `fetchRAGContext` → Python `POST /rag/context` (`domain_id`)
2. `buildRAGUserPrompt` + `config/prompts.json`
3. `callLLMCompletion` — OpenAI-compatible API
4. `cleanRAGAnswer`, `appendRAGDisclaimer`
5. `verifyRAGAnswer` — числа только из `fragments`

При fail verify — предупреждение пользователю, не «тихий» hallucination.

---

## `fetchRAGContext`

```json
POST PYTHON_RAG_URL
{ "question": "...", "domain_id": "default" }
```

Legacy: поле `crop_id`, env `CLASSIFIER_RAG_URL`.

---

## Deprecated

`POST /chat` — без сессии; заголовок `Deprecation: true`. Используйте `POST /message`.

---

## Логирование

`logRAGOutcome` — structured log `[RAG]` без тела LLM.

См. [quality-eval-and-rag-logs.md](./quality-eval-and-rag-logs.md).
