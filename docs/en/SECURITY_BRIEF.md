# Security Brief — Grounded LLM

**Audience:** IT security, procurement, compliance reviewers  
**Version:** Phase A (international product)  
**Last updated:** 2026

---

## Summary

Grounded LLM is an **on-premise or private-cloud** knowledge assistant. Employee questions are answered **only from documents you provide**. The platform does **not** train foundation models on your data.

| Property | Detail |
|----------|--------|
| Deployment | Docker Compose or Kubernetes (client infrastructure) |
| Data residency | Documents, chat history, and vector index stay in **your** environment |
| LLM usage | Optional external LLM API (OpenAI-compatible); prompts include retrieved excerpts only |
| Auth | Telegram Web App signature, API keys, admin Basic Auth |

---

## Data flow

```text
User (Web / Telegram / API)
        │
        ▼
   Go server ──► PostgreSQL (sessions, messages, feedback)
        │
        ├──► Python RAG service ──► Chroma (embeddings index)
        │         ▲
        │         └── reads documents from data/{tenant}/{domain}/
        │
        └──► LLM API (optional, HTTPS) — question + retrieved context only
```

**What leaves your network (if LLM API is external):**

- User question text
- Retrieved document **fragments** (chunks sent as context)
- System prompts configured in your locale bundles

**What does not leave:**

- Full document corpus (unless you choose to send it elsewhere)
- Vector database files
- Chat history (stored in your Postgres)

**What is never sent to the LLM vendor for model training:** Grounded LLM does not implement fine-tuning or training pipelines on client documents.

---

## Components and storage

| Component | Stores | Location |
|-----------|--------|----------|
| PostgreSQL | Users, sessions, messages, feedback, analytics events, **audit log** | Client DB volume |
| Chroma | Embedding vectors + chunk metadata | Client volume (`chroma_data`) |
| File system | Uploaded KB documents (`.txt`, `.pdf`, `.docx`) | `data/{tenant_id}/{domain_id}/` |
| Uploads | User images (optional) | Configured `UPLOAD_DIR` |

Multi-tenant isolation: `tenant_id` on sessions and in Chroma metadata filters.

---

## Authentication and access

| Surface | Mechanism |
|---------|-----------|
| Web App / chat API | Telegram `initData` HMAC (or `TELEGRAM_AUTH_DISABLED=true` for dev only) |
| Integrators | `X-API-Key` header |
| Admin (KB upload) | HTTP Basic Auth (`ADMIN_USER` / `ADMIN_PASSWORD`) |
| Python admin (reindex) | Shared secret `X-Admin-Secret` |

**Recommendations for production:**

- Set strong `ADMIN_PASSWORD` and `ADMIN_SECRET`
- Disable `TELEGRAM_AUTH_DISABLED`
- Restrict `CORS_ALLOWED_ORIGINS` to your domains
- Place admin and metrics endpoints behind VPN or reverse-proxy ACLs
- Rotate API keys periodically

---

## Subprocessors (optional)

If you configure an external LLM provider (e.g. OpenRouter, OpenAI, Azure OpenAI):

| Subprocessor | Purpose | Data shared |
|--------------|---------|-------------|
| Your chosen LLM API | Generate answers from retrieved context | Prompt (context + question) |

Review the LLM provider's DPA and data retention policy. For strict air-gap requirements, use a **local or VPC-hosted** OpenAI-compatible endpoint.

Embeddings model (`intfloat/multilingual-e5-small`) runs **inside the Python container** — no third-party embedding API by default.

---

## Logging and observability

- Structured request logs with `X-Request-ID`
- `[RAG]` logs: domain, session, fragment count, verify result — **no full LLM body**
- Prometheus metrics at `GET /metrics` (protect in production)
- **Admin audit log** (Postgres `audit_log`): failed admin login, successful admin verify (`GET /admin/status`), KB upload/delete/reindex — query via `GET /admin/audit-log` or Admin UI

---

## Hardening checklist (pilot → production)

- [ ] TLS termination at reverse proxy
- [ ] Secrets in vault / env injection (not committed `.env`)
- [ ] Postgres backups + restore tested
- [ ] Chroma + `data/` backup strategy
- [ ] Rate limits configured (`RATE_LIMIT_REQUESTS_PER_MINUTE`)
- [ ] LLM API key scoped and rotated
- [ ] Disable dev flags (`TELEGRAM_AUTH_DISABLED`)

---

## Contact

For security questionnaires during a pilot, request the architecture diagram and this brief together with your deployment diagram (network zones, egress to LLM).

See also: [ARCHITECTURE.md](./ARCHITECTURE.md), [DEPLOY.md](./DEPLOY.md), [PILOT_PLAYBOOK.md](./PILOT_PLAYBOOK.md).
