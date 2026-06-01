# RAG and LLM — `server/rag_pipeline.go`

**Sources:** `server/rag_pipeline.go`, `server/rag_chat.go`, `server/rag_verify.go`  
**Python:** [rag-retrieval.md](./rag-retrieval.md), [rag-verifier.md](./rag-verifier.md)  
**Called from:** `handleTextMessage` (`message_handlers.go`), `sse.go` (streaming)

---

## Pipeline

1. `fetchRAGContext` → Python `POST /rag/context` (`domain_id`, `tenant_id`, `locale`)
2. `buildRAGUserPrompt` + `config/locales/{locale}/prompts.json`
3. `callLLMCompletion` or `streamLLMCompletion` — OpenAI-compatible API
4. `cleanRAGAnswer`, `appendRAGDisclaimer`
5. `verifyRAGAnswer` — numbers must appear in `fragments`

On verify failure — warning to user, not a silent hallucination.

---

## `fetchRAGContext`

```json
POST PYTHON_RAG_URL
{
  "question": "...",
  "domain_id": "default",
  "tenant_id": "default",
  "locale": "en"
}
```

---

## Streaming

`POST /message?stream=1` — Server-Sent Events with incremental tokens.  
Web App uses streaming when supported, falls back to JSON response.

---

## Logging

`logRAGOutcome` — structured `[RAG]` log without LLM body.

See [quality-eval-and-rag-logs.md](./quality-eval-and-rag-logs.md).

---

## Citations

Assistant messages stored with `citations JSONB` (migration `005`).  
UI shows source excerpts linked to KB fragments.
