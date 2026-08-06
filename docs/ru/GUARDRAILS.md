**Канон (EN):** [GUARDRAILS.md](../en/GUARDRAILS.md)

# Интеграция Guardrails (опционально)

[grounded-guardrails](https://github.com/kantik001/grounded-guardrails) отдаёт `GuardrailsService` на **`:50052`**. Grounded LLM может вызывать его из пути verify после генерации LLM.

По умолчанию остаётся **in-process Spec verify** (`GUARDRAILS_MODE=local`) — CI и обычный `docker compose up` не меняются.

Зачем отдельный сервис: вынести token-level / PII-правила из ядра без ломки локального numeric verify.

## Место в архитектуре

```text
Retriever :50051 (python)     ← агенты / retrieval
Go server :8080               ← оркестрация + LLM
Guardrails :50052 (optional)  ← VerifyText после LLM
```

**Не** вешайте guardrails на `:50051` — порт занят Retriever.

## Режимы (`GUARDRAILS_MODE`)

| Режим | Поведение |
|-------|-----------|
| `local` (default) | In-process Spec numeric verify — CI и default Compose |
| `remote` | Только gRPC `VerifyText`; старт падает, если нет соединения |
| `hybrid` | Сначала gRPC; при транспортных ошибках — fallback на local |

Код клиента: `server/internal/guardrails/`. Локальный алгоритм: `server/internal/rag/verify.go`.

## Environment

```bash
GUARDRAILS_MODE=remote
GUARDRAILS_GRPC_ADDR=localhost:50052   # Compose: guardrails:50052
GUARDRAILS_PII_BLOCK=false             # также remote-правило pii_block
```

## Compose (sibling-репо)

```bash
# рядом: grounded-llm / grounded-guardrails
docker compose -f docker-compose.yml -f docker-compose.guardrails.yml up -d --build
```

## Вручную

```bash
# терминал 1 — guardrails
cd ../grounded-guardrails/go && go run ./cmd/server

# терминал 2 — grounded-llm:
# GUARDRAILS_MODE=remote GUARDRAILS_GRPC_ADDR=localhost:50052
```

## См. также

| Документ | Тема |
|----------|------|
| [ARCHITECTURE.md](./ARCHITECTURE.md) | Полный chat flow с verify |
| [NETWORK_SECURITY.md](./NETWORK_SECURITY.md) | Экспозиция `:50052` |
| [rag-verifier.md](./knowledge-base/rag-verifier.md) | Spec numeric algorithm |
| Proto | [`proto/guardrails.proto`](../../proto/guardrails.proto) |
