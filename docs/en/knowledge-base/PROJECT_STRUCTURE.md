# Grounded LLM — repository map

High-level map of the repository. Detailed articles: [README.md](./README.md).

---

## Root

| Path | Purpose |
|------|---------|
| `server/` | Go: auth, sessions, RAG+LLM, admin, verify |
| `api/` | Python Flask: RAG retrieval |
| `rag/` | Chroma, retrieval, domains, document loaders |
| `config/` | Domain pack defaults + `locales/{ru,en}/` |
| `data/{tenant}/{domain}/` | KB: `.txt`, `.pdf`, `.docx` |
| `webapp/` | Reference Telegram Web App UI |
| `migrations/` | PostgreSQL schema |
| `eval/`, `scripts/` | Quality & ops |
| `docs/` | Architecture, deploy, knowledge base (`en/`, `ru/`) |

---

## `server/` — Go backend

Single binary `package main`, module **`grounded_llm_server`**.

Key files:

| File | Role |
|------|------|
| `main.go`, `routes.go` | startup, routes |
| `domains.go`, `domain_resolve.go` | domain catalog, `domain_id` in API |
| `locale.go` | locale bundles, `X-Locale` middleware |
| `config_paths.go` | JSON lookup under `/config` |
| `rag_pipeline.go`, `rag_verify.go` | RAG + LLM + verify |
| `sse.go`, `llm_stream.go` | SSE streaming |
| `api_keys.go`, `auth_combined.go` | `X-API-Key`, `/api/v1/*` |
| `tenant.go` | `X-Tenant-ID`, KB path isolation |
| `postgres_store.go` | Postgres, sessions, messages |
| `admin.go` | KB upload, reindex |

→ [server-overview.md](./server-overview.md)

---

## `rag/` — RAG engine

| Module | Role |
|--------|------|
| `domains_config.py` | `config/domains.json` |
| `document_loaders.py` | `.txt`, `.pdf`, `.docx` |
| `vector_store.py` | Chroma indexing, tenant filter |
| `retrieval.py` | context for Go |
| `verifier.py` | number verify (mirror Go) |

---

## `config/` — domain pack

`domains.json`, `locales/{ru,en}/` (prompts, few_shot, onboarding, branding)

→ [config-overview.md](./config-overview.md)

---

## Documentation

| Path | Content |
|------|---------|
| `docs/en/ARCHITECTURE.md` | core vs domain pack |
| `docs/en/DEPLOY.md` | deployment |
| `docs/en/knowledge-base/` | module deep-dives |
| `docs/ru/` | Russian mirror |

---

## Outside core

Computer vision, industry-specific domain packs — **separate repos/packages**, not platform core.
