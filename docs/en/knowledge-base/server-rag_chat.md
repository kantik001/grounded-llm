# RAG and LLM — `server/rag_pipeline.go`

**Sources:** `server/rag_pipeline.go`, `server/rag_chat.go`, `server/rag_verify.go`, `server/guardrails_client.go`  
**Python:** [rag-retrieval.md](./rag-retrieval.md), [rag-verifier.md](./rag-verifier.md)  
**Remote verify (optional):** [../GUARDRAILS.md](../GUARDRAILS.md)  
**Called from:** `handleTextMessage` (`message_handlers.go`), `sse.go` (streaming)

---

## Pipeline

1. `fetchRAGContext` → Python `POST /rag/context` (`domain_id`, `tenant_id`, `locale`) + `X-Request-ID`
2. `buildRAGUserPrompt` + `config/locales/{locale}/prompts.json`
3. `callLLMCompletion` or `streamLLMCompletion` — OpenAI-compatible API
4. `cleanRAGAnswer`, `appendRAGDisclaimer`
5. `verifyRAGAnswer` — Spec numeric check (±0.01 vs fragments):
   - **`GUARDRAILS_MODE=local` (default):** in-process extract/compare in `rag_verify.go`
   - **`remote` / `hybrid`:** gRPC `VerifyText` → [grounded-guardrails](https://github.com/kantik001/grounded-guardrails) `:50052` (`guardrails_client.go`); hybrid falls back to local on transport errors

On verify failure — warning to user, not a silent hallucination.

---

## Request path trace (MVP)

Same `X-Request-ID` / `request_id` across the chat path for debug/tuning (`server/request_path_trace.go`):

| Step log | Meaning |
|----------|---------|
| `step=message.accept` | `POST /message` accepted |
| `step=cache` | response cache hit/miss |
| `step=retrieve` | Python `/rag/context` (Go + Python logs) |
| `step=llm` | completion (stream true/false) |
| `step=verify` | local/remote/hybrid + pass |
| `step=done` | total `ms`, outcome |

Client sees `request_id` in JSON / SSE `done` (and `X-Request-ID` response header). Grep logs: `req=<id>`.

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

Verify still runs on the **final** assembled answer (same `finalizeRAGAnswer` path).

---

## Logging

Path-trace steps (`req=… step=…`) plus `logRAGOutcome` — structured `[RAG]` log without LLM body.

See [quality-eval-and-rag-logs.md](./quality-eval-and-rag-logs.md).

---

## Citations

Assistant messages stored with `citations JSONB` (migration `005`).  
UI shows source excerpts linked to KB fragments.
