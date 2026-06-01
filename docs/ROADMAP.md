# Roadmap — Grounded LLM

## Done (platform core)

- Document pipeline: `.txt`, `.pdf`, `.docx`
- Admin upload + reindex
- Legacy API removed (`crop_id`, `/crops`, …)
- **Phase 1:** citations in UI, `schema_migrations`, expanded eval baseline, admin index stats + delete, `rag_k` per domain, removed `POST /chat`

## Phase 2 — Integrators (next)

- Streaming (SSE) in Go + Web App
- API keys (`X-API-Key`), OpenAPI
- Multi-tenancy (`tenant_id`, isolated Chroma collections)
- Domain pack template + CLI

## Phase 3 — Platform & monetization

- Helm / Terraform, managed vector DB
- Open core vs hosted SaaS
- Optional vision domain pack, audit log, analytics dashboard

See also: [ARCHITECTURE.md](./ARCHITECTURE.md), [DEPLOY.md](./DEPLOY.md).
