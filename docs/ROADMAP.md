# Roadmap — Grounded LLM

## Done

### Platform core
- Document pipeline: `.txt`, `.pdf`, `.docx`
- Admin upload + reindex, citations UI, eval baseline
- Legacy API removed, `schema_migrations`

### Phase 1 — Trust
- Citations in chat, `rag_k`, admin index stats + delete, expanded eval CI

### Phase 2 — Integrators
- **SSE streaming** — `POST /message?stream=1` (Web App uses stream with JSON fallback)
- **API keys** — `X-API-Key`, env `API_KEYS` or `API_KEYS_FILE`
- **API v1** — `/api/v1/*` + `GET /api/v1/openapi.json`
- **Multi-tenant (minimal)** — `X-Tenant-ID`, `data/{tenant}/{domain}/`, Chroma filter `tenant_id`
- **Observability** — `X-Request-ID`, `GET /metrics` (Prometheus text), structured request logs
- **Admin feedback** — `GET /admin/feedback`
- **Domain scaffold** — `scripts/init_domain.sh` / `init_domain.ps1`

## Phase 3 — Platform & monetization (next)

- Helm / Terraform, managed vector DB
- Open core vs hosted SaaS
- Vision domain pack, audit log, analytics dashboard

See also: [ARCHITECTURE.md](./ARCHITECTURE.md), [DEPLOY.md](./DEPLOY.md).
