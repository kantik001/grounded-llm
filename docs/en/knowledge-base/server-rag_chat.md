# RAG and LLM — Go orchestration

**Sources:** `internal/rag/pipeline.go`, `internal/rag/verify.go`, `internal/guardrails/client.go`, `internal/httpapi/message.go`, `internal/httpapi/sse.go`  
**Python:** [rag-retrieval.md](./rag-retrieval.md), [rag-verifier.md](./rag-verifier.md)  
**Remote verify (optional):** [../GUARDRAILS.md](../GUARDRAILS.md)  
**Entry:** `POST /message` / `POST /message?stream=1` in `internal/httpapi`

---

## Pipeline

1. `fetchRAGContext` → Python `POST /rag/context` (`domain_id`, `tenant_id`, `locale`) + `X-Request-ID`
2. `buildRAGUserPrompt` + `config/locales/{locale}/prompts.json`
3. `callLLMCompletion` or `streamLLMCompletion` — OpenAI-compatible API (`internal/llm`)
4. `CleanAnswer`, `AppendDisclaimer`
5. `verifyRAGAnswer` — Spec numeric check (±0.01 vs fragments), optional faithfulness / NLI:
   - **`GUARDRAILS_MODE=local` (default):** in-process in `internal/rag/verify.go`
   - **`remote` / `hybrid`:** gRPC `VerifyText` → [grounded-guardrails](https://github.com/kantik001/grounded-guardrails) `:50052`; hybrid falls back to local on transport errors
   - **`VERIFY_FAITHFULNESS=warn|enforce`:** lexical sentence support (`verify_faithfulness.go`)
   - **`VERIFY_NLI=assist|replace`:** optional HTTP NLI (`verify_nli.go`)

On verify failure — warning to user, not a silent hallucination.

---

## Request path trace (MVP)

Same `X-Request-ID` / `request_id` across the chat path (`internal/httpapi/path_trace.go`):

| Step log | Meaning |
|----------|---------|
| `step=message.accept` | `POST /message` accepted |
| `step=cache` | response cache hit/miss |
| `step=retrieve` | Python `/rag/context` |
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

Verify still runs on the **final** assembled answer (same finalize path).

---

## Logging

Path-trace steps (`req=… step=…`) plus structured `[RAG]` log without LLM body (`internal/rag/log.go`).

See [quality-eval-and-rag-logs.md](./quality-eval-and-rag-logs.md).

---

## Citations

Assistant messages stored with `citations JSONB` (migration `005`).  
UI shows source excerpts linked to KB fragments.
