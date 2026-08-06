Связанный EN-канон (фрагментарно): [SAAS.md](../../en/SAAS.md) · [ANALYTICS_GUIDE.md](../../en/ANALYTICS_GUIDE.md) · [server-overview.md](../../en/knowledge-base/server-overview.md) · [TRUST_CENTER.md](../../en/TRUST_CENTER.md)

# Go: OIDC, SaaS, analytics, audit, metrics

Deep dive по пакетам после **strangler extract**. Entry point: `server/cmd/server` → `internal/app.Run()`. Плоских `server/*.go` больше нет — только `internal/*`.

Зачем эта статья: связать «опциональный» слой (SSO, signup, дашборд качества, Prometheus) с реальными путями в коде и миграциями `010` / `011`.

Порты (напоминание): Go **8080**, Python HTTP **5000**, Retriever gRPC **50051**, guardrails **50052**, Redis **6379**.

---

## Карта пакетов

| Пакет | Путь | Роль |
|-------|------|------|
| OIDC SSO | `server/internal/oidc/` | Login / callback / session cookie для admin |
| SaaS | `server/internal/saas/` | Signup, plans, Stripe Checkout + webhook |
| Analytics | `server/internal/analytics/` | Запись событий RAG в store |
| Audit | `server/internal/audit/` | Admin audit log (кто что сделал) |
| Metrics | `server/internal/metrics/` | Process-wide counters для `/metrics` |
| Store | `server/internal/store/` | Postgres: analytics dashboard, audit, SaaS tables, purge |
| HTTP scrape | `server/internal/httpapi/metrics_handler.go` | `GET /metrics` |
| Admin routes | `server/internal/admin/` | Analytics UI API, tenant purge, RBAC |

Composition: `internal/app/*_bridge.go` прокидывает deps (store, config) в пакеты через `bind.go`.

```text
cmd/server
    └── internal/app (Run, Deps)
            ├── oidc.LoadSettings + RegisterAuthRoutes
            ├── saas.RegisterRoutes
            ├── analytics.RecordRAG (из rag pipeline)
            ├── audit.Record (admin / OIDC)
            └── httpapi → metrics scrape + chat
```

---

## `internal/oidc` — SSO для админки

**Зачем:** не раздавать shared Basic-пароль админам; вход через корпоративный IdP (Okta, Keycloak, Azure AD…).

| Файл | Назначение |
|------|------------|
| `doc.go` | Package doc |
| `config.go` | Env: `OIDC_ENABLED`, `OIDC_ISSUER`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`, `OIDC_REDIRECT_URL`, `OIDC_SCOPES`, `OIDC_SESSION_SECRET`, `OIDC_SESSION_TTL_HOURS` + role mapping |
| `provider.go` | Discovery OIDC provider + oauth2 config |
| `handlers.go` | `GET /config`, `/login`, `/callback`, `POST /logout` |
| `session.go` | Signed cookie `grounded_admin_session` (`SessionPayload`: sub, email, roles, exp, auth=`oidc`) |
| `bind.go` | Wiring config/store/audit |

Маршруты монтируются под admin auth group (типично `/api/admin/auth/*`). При ошибке OIDC — audit `login_failed`; при успехе — `login`.

Роли по умолчанию: editor → `kb_editor`, admin → `admin` (см. mapping в `config.go`). Без `OIDC_SESSION_SECRET` можно опереться на `ADMIN_SECRET` для подписи cookie.

Подробнее по env: `config/SSO.md`.

---

## `internal/saas` — signup и биллинг

**Зачем:** optional hosted-слой поверх того же ядра (on-prem не обязан включать).

| Файл | Назначение |
|------|------------|
| `signup.go` | `POST /api/v1/signup`, `GET /api/v1/plans` |
| `plans.go` | Каталог из `PLANS_FILE` (default `config/plans.yaml`) |
| `stripe_checkout.go` | Checkout session для платных планов |
| `stripe_webhook.go` | `POST /api/v1/billing/stripe/webhook` + idempotency |
| `bind.go` | Tenant registry, quotas, admin users |

Включение: `SAAS_SIGNUP_ENABLED=true` + настроенный tenants registry.

Поток signup (кратко):

1. Валидация `org_name` / `email` / `plan`  
2. Создание tenant + квот + `data/{tenant}/`  
3. Опционально admin user (`{tenant}-admin`)  
4. Платный план → `checkout_url` при Stripe  

### Миграция `010_saas_tenants.sql`

Таблицы:

- `saas_tenants` — tenant_id, org, email, plan, stripe_customer_id  
- `tenant_quotas` — messages_per_day, storage_mb, max_domains  
- `stripe_webhook_events` — idempotency по `event_id`  

### Миграция `011_admin_users_membership.sql`

- `admin_users` — username, password_bcrypt, roles[], tenant_id  
- `user_tenant_memberships` — telegram_id × tenant_id × role  

Русский обзор SaaS: [../SAAS.md](../SAAS.md). Биллинг: [../BILLING.md](../BILLING.md).

Store: `server/internal/store/saas_store.go`, `admin_membership_store.go`.

---

## `internal/analytics` — качество ассистента

**Зачем:** отличить «плохо отвечает, потому что нет дока» от «галлюцинирует числа» без чтения всех логов.

| Файл | Назначение |
|------|------------|
| `recorder.go` | `RecordRAG` → событие `rag_answer` |
| `bind.go` | Store wiring |
| `analytics_test.go` | Unit tests |

`RecordRAG` пишет payload:

- `tenant_id`, `domain_id`  
- `verify_pass`, `soft_fail`, `fragment_count`  
- `question_preview` — до **80** символов  

Запись при soft-fail или при успешном `OK` без `ErrMsg`. Вызов — из RAG/chat pipeline после verify.

### Дашборд (store)

`ChatStore.AnalyticsDashboard` (`store/analytics_dashboard.go`, `analytics_store.go`):

| Поле | Смысл |
|------|-------|
| `questions_total` / `questions_today` | Объём |
| `rag.verify_pass_rate` | Доля прошедших numeric verify |
| `rag.soft_fail` | «Нет в KB» / слабый retrieval |
| `kb_gaps` | Превью вопросов без ответа |
| `feedback` | Thumbs |
| `top_domains` | Нагрузка по domain_id |

Окно: 1–90 дней (default 7). UI: `webapp/admin.html` → Analytics.

Продуктовые решения: [../ANALYTICS_GUIDE.md](../ANALYTICS_GUIDE.md).

---

## `internal/audit` — кто менял систему

**Зачем:** compliance / post-incident — не путать с product analytics.

| Файл | Назначение |
|------|------------|
| `record.go` | `audit.Record(c, Opts{Action, Actor, TenantID, …})` |
| `bind.go` | Store |

Пишет в Postgres через `store.AuditRecord` (IP из `X-Forwarded-For` / `X-Real-IP`, `X-Request-ID`). Типичные actions: `login`, `login_failed`, `logout`, `tenant_purge`, admin ops (reindex, upload…).

Purge tenant: `DELETE /api/admin/tenants/:id` — [../TENANT_PURGE.md](../TENANT_PURGE.md); handler в `internal/admin/handlers_ops.go`, SQL в `internal/store/tenant_purge_store.go`. Перед wipe — audit row `tenant_purge`.

---

## `internal/metrics` — Prometheus counters

**Зачем:** capacity и latency без захода в admin UI; scrape из K8s / Grafana.

| Символ | Назначение |
|--------|------------|
| `HTTPRequests`, `RAGRequests`, `LLMRequests` | Объём |
| `CacheHits` / `CacheMisses` | Response cache (Go / Redis) |
| `RecordLLMUsage` | Tokens + latency + TTFT per tenant/model |
| `EstimateTokens` | ~4 chars/token, если нет `usage` от провайдера |

Экспорт: `internal/httpapi/metrics_handler.go` → `GET /metrics` на **:8080**.

В production без `METRICS_TOKEN` endpoint **закрыт** (видны per-tenant counters). С токеном — `Authorization: Bearer …`.

Связанные серии (имена в scrape): `llm_tokens_input_total`, `llm_tokens_output_total`, `llm_latency_seconds_*`, `llm_ttft_seconds_*`, cache counters. Python отдельно отдаёт embedding-cache метрики на своём `/metrics`.

LLM/кэши: [../LLM_PROVIDERS.md](../LLM_PROVIDERS.md).

---

## Как это стыкуется в chat flow

```text
Client → Go :8080  POST /message
    → (опц.) Redis response cache
    → Python :5000  POST /rag/context
    → LLM (OpenAI-compatible)
    → verify (local | remote :50052 | hybrid)
    → analytics.RecordRAG
    → Postgres messages + citations
    → metrics.RecordLLMUsage
```

Admin с OIDC → audit login → Analytics dashboard читает те же Postgres events.

Eval качества retrieval (**99** cases) — отдельно в CI (`eval-retrieval-gate`), не через analytics package. См. [../BENCHMARK.md](../BENCHMARK.md).

---

## Порядок чтения

1. [server-overview.md](./server-overview.md) — карта `internal/*`  
2. [server-auth-and-limits.md](./server-auth-and-limits.md) — API keys / Telegram  
3. Эта статья — OIDC / SaaS / analytics / audit / metrics  
4. [server-rag_chat.md](./server-rag_chat.md) — где вызывается RecordRAG + verify  
5. [migrations-overview.md](./migrations-overview.md) — `010`, `011`  

---

## См. также

- [../SAAS.md](../SAAS.md) · [../BILLING.md](../BILLING.md)  
- [../NETWORK_SECURITY.md](../NETWORK_SECURITY.md) — не публиковать `:5000` / `:50051` / `:50052` / Redis  
- [../COMPATIBILITY.md](../COMPATIBILITY.md)  
