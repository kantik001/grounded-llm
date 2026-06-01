# Roadmap — Grounded LLM

## Done

### Platform core
- Document pipeline: `.txt`, `.pdf`, `.docx`
- Admin upload + reindex, citations UI, eval baseline
- Legacy API removed, `schema_migrations`

### Phase 1 — Trust
- Citations in chat, `rag_k`, admin index stats + delete, expanded eval CI

### Phase 2 — Integrators
- **SSE streaming** — `POST /message?stream=1`
- **API keys** — `X-API-Key`, env `API_KEYS` or `API_KEYS_FILE`
- **API v1** — `/api/v1/*` + `GET /api/v1/openapi.json`
- **Multi-tenant (minimal)** — `X-Tenant-ID`, `data/{tenant}/{domain}/`, Chroma filter `tenant_id`
- **Observability** — `X-Request-ID`, `GET /metrics`, structured request logs
- **Admin feedback** — `GET /admin/feedback`
- **Domain scaffold** — `scripts/init_domain.sh` / `init_domain.ps1`

### i18n
- Docs `docs/en/` and `docs/ru/`
- Locale bundles `config/locales/{ru,en}/`
- Middleware `X-Locale`, `Accept-Language`, `?locale=`

## Phase 3 — Platform & monetization (next)

- Helm / Terraform, managed vector DB
- Open core vs hosted SaaS
- Vision domain pack, audit log, analytics dashboard

See also: [ARCHITECTURE.md](./ARCHITECTURE.md), [DEPLOY.md](./DEPLOY.md).
