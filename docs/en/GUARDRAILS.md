# Guardrails integration (optional)

[grounded-guardrails](https://github.com/kantik001/grounded-guardrails) exposes `GuardrailsService` on **`:50052`**. Grounded LLM can call it from the RAG verify path instead of (or before) the in-process numeric check.

## Modes (`GUARDRAILS_MODE`)

| Mode | Behavior |
|------|----------|
| `local` (default) | In-process Spec verify — CI and default Compose unchanged |
| `remote` | gRPC `VerifyText` only; startup fails if dial fails |
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

Proto (synced from guardrails): [`api/proto/guardrails.proto`](../../api/proto/guardrails.proto).
