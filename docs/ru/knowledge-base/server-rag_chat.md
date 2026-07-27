# RAG и LLM — `server/rag_pipeline.go`

**Исходники:** `server/rag_pipeline.go`, `server/rag_chat.go`, `server/rag_verify.go`, `server/guardrails_client.go`  
**Python:** [rag-retrieval.md](./rag-retrieval.md), [rag-verifier.md](./rag-verifier.md)  
**Удалённый verify (опц.):** [../en/GUARDRAILS.md](../en/GUARDRAILS.md)  
**Вызывается из:** `handleTextMessage`, `sse.go` (streaming)

---

## Цепочка

1. `fetchRAGContext` → Python `POST /rag/context` (`domain_id`, `tenant_id`, `locale`)
2. `buildRAGUserPrompt` + `config/locales/{locale}/prompts.json`
3. `callLLMCompletion` или `streamLLMCompletion` — OpenAI-compatible API
4. `cleanRAGAnswer`, `appendRAGDisclaimer`
5. `verifyRAGAnswer` — Spec-проверка чисел (±0.01 относительно фрагментов):
   - **`GUARDRAILS_MODE=local` (по умолчанию):** in-process в `rag_verify.go`
   - **`remote` / `hybrid`:** gRPC `VerifyText` → [grounded-guardrails](https://github.com/kantik001/grounded-guardrails) `:50052` (`guardrails_client.go`); hybrid при сбое сети откатывается на local

При неудачной verify — предупреждение пользователю, а не «тихая» галлюцинация.

---

## `fetchRAGContext`

```json
POST PYTHON_RAG_URL
{
  "question": "...",
  "domain_id": "default",
  "tenant_id": "default",
  "locale": "ru"
}
```

---

## Потоковый ответ

`POST /message?stream=1` — Server-Sent Events с пошаговой выдачей текста.

Verify выполняется на **финальном** собранном ответе (тот же `finalizeRAGAnswer`).

---

## Citations

Сообщения ассистента сохраняются с `citations JSONB` (миграция `005`). UI показывает выдержки из KB.

---

## Логирование

`logRAGOutcome` — структурированный лог `[RAG]` без тела LLM.

→ [quality-eval-and-rag-logs.md](./quality-eval-and-rag-logs.md)
