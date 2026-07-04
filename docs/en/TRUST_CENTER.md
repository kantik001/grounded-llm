# Trust center

Summary for security reviewers and procurement questionnaires. Detailed controls live in linked docs.

## Product summary

**Grounded LLM** is an on-prem / private-cloud platform for document-grounded assistants with citations and numeric verify. Core stack: Go API, Python RAG (Chroma), Postgres, optional Telegram Web App UI.

## Data handling

| Data type | Stored where | Encrypted at rest | Retention |
|-----------|--------------|-------------------|-----------|
| Chat messages | Postgres | Operator responsibility (disk/DB encryption) | Configurable via `MESSAGE_RETENTION_DAYS` |
| Chat sessions | Postgres | Same | `SESSION_RETENTION_DAYS` |
| Documents (KB) | `data/` volume | Operator responsibility | Until admin delete |
| Embeddings | Chroma volume | Operator responsibility | Rebuilt from KB |
| User images | `UPLOAD_DIR` | File permissions `0600` | With session/message retention |
| Audit events | Postgres | Same as DB | Follow DB backup policy |

LLM prompts may include retrieved context and recent chat history. Review your LLM provider DPA; use VPC-hosted endpoints for strict data residency.

## Authentication and authorization

- **End users:** Telegram Web App init data (HMAC) or `X-API-Key` for programmatic access
- **Admin UI:** HTTP basic + optional OIDC SSO; RBAC roles (see `config/RBAC.md`)
- **Internal:** `RAG_SERVICE_TOKEN` between Go and Python; `ADMIN_SECRET` for Python admin routes

## Network security

- Python RAG is not intended for public exposure; place on internal network only
- See [NETWORK_SECURITY.md](./NETWORK_SECURITY.md) for nginx CSP, CORS, and ingress guidance

## Observability and audit

- Structured logs with `X-Request-ID`
- Prometheus-style metrics at `/metrics`
- Audit log for admin actions (upload, delete, reindex, login) — see `config/AUDIT.md`

## Subprocessors

| Component | Purpose | Operator choice |
|-----------|---------|-----------------|
| LLM API (OpenRouter, OpenAI, etc.) | Answer generation | Configurable `LLM_BASE_URL` |
| HuggingFace (embeddings) | Local model download on first Python start | Can mirror models internally |

No mandatory SaaS beyond what the operator configures.

## Vulnerability reporting

See [SECURITY.md](../../SECURITY.md) in the repository root.

## Compliance posture

The platform provides **technical controls** (auth, audit, retention, on-prem deploy). Formal certifications (SOC 2, ISO 27001) are the operator's responsibility on their deployment.

## Related documentation

- [SECURITY_BRIEF.md](./SECURITY_BRIEF.md)
- [ARCHITECTURE.md](./ARCHITECTURE.md)
- [BACKUP_RESTORE.md](./BACKUP_RESTORE.md)
- [K8S_DEPLOY.md](./K8S_DEPLOY.md)
