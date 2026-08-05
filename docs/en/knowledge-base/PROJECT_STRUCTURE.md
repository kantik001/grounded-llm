# Grounded LLM — repository map

High-level map of the repository. Detailed articles: [README.md](./README.md).

---

## Root

| Path | Purpose |
|------|---------|
| `server/` | Go: auth, sessions, RAG+LLM, admin, verify — [`server/README.md`](../../../server/README.md) (`cmd/server` + `internal/`) |
| `api/` | Python RAG **service** (internal HTTP+gRPC); see `api/README.md` |
| `proto/` | Guardrails gRPC IDL (`:50052`); Retriever in `api/proto/` — see `proto/README.md`, root `buf.yaml` |
| `rag/` | Retrieval **engine** (library); HTTP/gRPC in `api/` — see [`rag/README.md`](../../../rag/README.md) |
| `config/` | Domain pack defaults + `locales/{ru,en}/` |
| `data/{tenant}/{domain}/` | KB: `.txt`, `.pdf`, `.docx` |
| `webapp/` | Reference UI — [`webapp/README.md`](../../../webapp/README.md) |
| `site/` | GitHub Pages landing — [`site/README.md`](../../../site/README.md) |
| `migrations/` | PostgreSQL schema — see [`migrations/README.md`](../../../migrations/README.md) |
| `eval/`, `scripts/`, `tests/` | Quality & ops — READMEs under each folder |
| `sdk/python/` | Python client + CLI for Go API — [`sdk/README.md`](../../../sdk/README.md) |
| `models/`, `sparse_index/` | Runtime binaries/caches (gitignored) — see folder READMEs |
| `docs/` | Architecture, deploy, knowledge base (`en/`, `ru/`) |

---

## `server/` — Go backend

Module **`grounded_llm_server`**: `cmd/server` + `internal/` (see [`server/README.md`](../../../server/README.md)).

| Path | Role |
|------|------|
| `cmd/server` | process entrypoint |
| `internal/app` | composition: `Run()`, `Deps`, bridges |
| `internal/{config,store,auth,guardrails,metrics,llm,rag,httpapi,locale,domain,tenant,admin,oidc,saas,audit,analytics}` | domain packages |
| `gen/guardrails/v1` | guardrails gRPC stubs |

→ [server-overview.md](./server-overview.md)

---

## `rag/` — RAG engine

Library used by `api/` and scripts — not a network service. Full map + env: [`rag/README.md`](../../../rag/README.md).

| Module | Role |
|--------|------|
| `vector_backend/` | Chroma / Qdrant / pgvector |
| `vector_store.py` | search, hybrid RRF, readiness |
| `indexing.py` / `document_loaders.py` | load + chunk KB files |
| `retrieval.py` | context + few-shot for callers |
| `sparse_index.py` / `rerank.py` | BM25 hybrid + optional rerank |
| `verifier.py` | local numeric check |

---

## `config/` — domain pack

`domains.json`, `locales/{ru,en}/`, `examples/`, `schemas/` — see [config/README.md](../../../config/README.md)

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
