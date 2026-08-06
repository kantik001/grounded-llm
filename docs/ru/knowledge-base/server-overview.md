# Go-сервер — обзор

**Папка:** [`server/`](../../../server/README.md)  
**Роль:** оркестратор — авторизация, API, PostgreSQL, Python RAG, LLM, verify  
**Фреймворк:** [Gin](https://gin-gonic.com/)  
**Порт:** `8080`  
**Сборка:** `cd server && go build -o main ./cmd/server`

| Документ | Тема |
|----------|------|
| [server-auth-and-limits.md](./server-auth-and-limits.md) | Telegram, API keys, CORS, лимиты |
| [server-chat-and-db.md](./server-chat-and-db.md) | Чат, БД, сессии |
| [server-rag_chat.md](./server-rag_chat.md) | RAG + LLM + streaming |
| [server-admin-and-ux-api.md](./server-admin-and-ux-api.md) | Админка, domains, onboarding |

---

## Пакеты (сейчас)

Зрелый Go layout — strangler extract **завершён**. См. [`server/README.md`](../../../server/README.md).

| Путь | Роль |
|------|------|
| `cmd/server` | `main()` → `app.Run()` |
| `internal/app` | composition root: `Run()`, `Deps`, тонкие `*_bridge.go` |
| `internal/config` | `Config`, load/validate |
| `internal/store` | `ChatStore`, Postgres, retention |
| `internal/auth` | Telegram initData, API keys, HTTP auth middleware |
| `internal/guardrails` | gRPC-клиент remote/hybrid verify |
| `internal/metrics` | общие Prometheus-счётчики |
| `internal/llm` | OpenAI-compatible chat + stream/cache/mock |
| `internal/rag` | retrieve → prepare → verify → answer |
| `internal/httpapi` | routes, CORS, health/chat/SSE, rate limit, OpenAPI |
| `internal/locale` | locale bundles, branding, onboarding |
| `internal/domain` | каталог доменов + RAG guards |
| `internal/tenant` | tenants, quotas, registry |
| `internal/admin` | admin HTTP + RBAC + reindex |
| `internal/oidc` | OIDC SSO |
| `internal/saas` | signup, plans, Stripe |
| `internal/audit` | audit helpers |
| `internal/analytics` | RAG analytics recorder |
| `gen/guardrails/v1` | stubs для `:50052` |

**Vision/CV** — вне ядра; подключается domain pack при необходимости.

---

## Схема сервисов

```mermaid
flowchart TB
    Web[webapp / Telegram / API clients]
    Agents[Agents gRPC]
    Go[server Go :8080]
    PG[(PostgreSQL)]
    Redis[(Redis caches)]
    Py[python RAG HTTP :5000 + gRPC :50051]
    LLM[OpenAI-compatible LLM\nOpenRouter / Ollama / vLLM]
    GR[grounded-guardrails :50052\noptional]

    Web --> Go
    Agents -->|gRPC :50051| Py
    Go --> PG
    Go --> Redis
    Go --> Py
    Go --> LLM
    Go --> GR
```

См. также architecture docs в `docs/ru/`.
