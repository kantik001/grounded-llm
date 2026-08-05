# Go server (`server/`)

Public HTTP API orchestrator (Gin, `:8080`): auth → Python RAG → LLM → verify.

## Layout

```text
server/
  cmd/server/                 # main() → app.Run()
  internal/
    app/                      # composition root: Run(), bridges, wiring
    config/                   # Config load/validate
    store/                    # ChatStore + Postgres + retention
    auth/                     # Telegram initData, API keys, auth middleware
    guardrails/               # remote/hybrid gRPC verify client
    metrics/                  # shared Prometheus counters
    llm/                      # OpenAI-compatible chat + stream/cache/mock
    rag/                      # retrieve → prepare → verify → answer
    httpapi/                  # routes, CORS, health/chat/SSE, rate limit
    locale/                   # locale bundles, branding, onboarding prompts
    domain/                   # domain catalog + guards
    tenant/                   # tenants, quotas, registry
    admin/                    # admin HTTP + RBAC + reindex
    oidc/                     # OIDC SSO
    saas/                     # signup, plans, Stripe
    audit/                    # audit log helpers
    analytics/                # RAG analytics recorder
  gen/guardrails/v1/          # generated from proto/guardrails.proto
  go.mod
```

**vs `api/` / `rag/`:** this binary is the product API. Retrieval engine is `rag/`; Retriever HTTP/gRPC service is `api/`.

### Dependency rule

```text
cmd/server → internal/app (Deps + bridges) → leaf packages
```

- Prefer `app.D` (`Deps`) over new bare globals.
- **Extracted:** config, store, auth, guardrails, metrics, llm, rag, httpapi, locale, domain, tenant, admin, oidc, saas, audit, analytics.
- `app` keeps thin `*_bridge.go` wrappers and `Run()` so call sites stay stable.

### `app.Deps`

```go
type Deps struct {
    Config *config.Config
    Store  *store.ChatStore
}
```

## Build / test

```bash
cd server
go build -o main ./cmd/server
go test ./...
```

Docker: [`Dockerfile.server`](../Dockerfile.server) builds `./cmd/server`.

## Docs

- [server-overview.md](../docs/en/knowledge-base/server-overview.md)
- [GUARDRAILS.md](../docs/en/GUARDRAILS.md)
- Guardrails IDL: [`proto/guardrails.proto`](../proto/guardrails.proto)
