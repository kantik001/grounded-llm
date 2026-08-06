# RAG и LLM — оркестрация Go

**Исходники:** `internal/rag/pipeline.go`, `internal/rag/verify.go`, `internal/guardrails/client.go`, `internal/httpapi/message.go`, `internal/httpapi/sse.go`  
**Python:** [rag-retrieval.md](./rag-retrieval.md), [rag-verifier.md](./rag-verifier.md)  
**Remote verify:** [../GUARDRAILS.md](../GUARDRAILS.md)  
**EN:** [server-rag_chat.md](../../en/knowledge-base/server-rag_chat.md)

---

## Конвейер

1. `fetchRAGContext` → Python `POST /rag/context` (`domain_id`, `tenant_id`, `locale`) + `X-Request-ID`
2. Промпт из `config/locales/{locale}/prompts.json`
3. LLM (`internal/llm`) — OpenAI-compatible (OpenRouter / Ollama / vLLM)
4. `CleanAnswer`, `AppendDisclaimer`
5. `VerifyRAGAnswer`:
   - **`GUARDRAILS_MODE=local`** (default): числа ±0.01 в `internal/rag/verify.go`
   - **`remote` / `hybrid`:** gRPC `VerifyText` → grounded-guardrails `:50052`
   - **`VERIFY_FAITHFULNESS=warn|enforce`:** лексическая поддержка предложений
   - **`VERIFY_NLI=assist|replace`:** опциональный HTTP NLI

При fail verify — предупреждение пользователю, не «тихая» галлюцинация.

---

## Request path trace

Один `X-Request-ID` / `request_id` на весь путь (`internal/httpapi/path_trace.go`):

| Step | Смысл |
|------|--------|
| `message.accept` | принят `POST /message` |
| `cache` | hit/miss response cache |
| `retrieve` | Python `/rag/context` |
| `llm` | completion |
| `verify` | local/remote/hybrid |
| `done` | итог `ms` |

Клиент видит `request_id` в JSON / SSE `done`.

---

## Streaming

`POST /message?stream=1` — SSE. Verify на **финальном** собранном ответе.

---

## Citations

`citations JSONB` (миграция `005`). UI показывает фрагменты источников.

---

## Дальше

| Тема | Файл |
|------|------|
| Verify подробно | [rag-verifier.md](./rag-verifier.md) |
| Логи / eval | [quality-eval-and-rag-logs.md](./quality-eval-and-rag-logs.md) |
| LLM providers | [../LLM_PROVIDERS.md](../LLM_PROVIDERS.md) |
