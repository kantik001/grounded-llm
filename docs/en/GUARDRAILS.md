# Guardrails integration (optional)

[grounded-guardrails](https://github.com/kantik001/grounded-guardrails) exposes `GuardrailsService` on **`:50052`**. Grounded LLM can call it from the RAG verify path after generation.

Default remains **in-process Spec verify** (`GUARDRAILS_MODE=local`) so CI and plain `docker compose up` stay unchanged.

## Architecture placement

```text
Retriever :50051 (python)     ← agents / retrieval
Go server :8080               ← orchestration + LLM
Guardrails :50052 (optional)  ← VerifyText after LLM
```

Do **not** bind guardrails on `:50051` — that port is owned by the Retriever.

## Modes (`GUARDRAILS_MODE`)

| Mode | Behavior |
|------|----------|
| `local` (default) | In-process Spec numeric verify — CI and default Compose |
| `remote` | gRPC `VerifyText` only; startup fails if connect fails |
| `hybrid` | Prefer gRPC; fall back to local on transport errors |

## Environment

```bash
GUARDRAILS_MODE=remote
GUARDRAILS_GRPC_ADDR=localhost:50052   # Compose: guardrails:50052
GUARDRAILS_PII_BLOCK=false             # also run pii_block rule remotely
```

## Compose (sibling repo)

```bash
# clones side by side: grounded-llm / grounded-guardrails
docker compose -f docker-compose.yml -f docker-compose.guardrails.yml up -d --build
```

## Manual

```bash
# terminal 1 — guardrails
cd ../grounded-guardrails/go && go run ./cmd/server

# terminal 2 — grounded-llm server with:
# GUARDRAILS_MODE=remote GUARDRAILS_GRPC_ADDR=localhost:50052
```

## Related

| Doc | Topic |
|-----|--------|
| [ARCHITECTURE.md](./ARCHITECTURE.md) | Full chat flow including verify step |
| [NETWORK_SECURITY.md](./NETWORK_SECURITY.md) | Port exposure for `:50052` |
| [rag-verifier.md](./knowledge-base/rag-verifier.md) | Spec numeric algorithm |
| Proto | [`proto/guardrails.proto`](../../proto/guardrails.proto) |
