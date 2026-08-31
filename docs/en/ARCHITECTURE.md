# Architecture: Grounded LLM

This repository is the **platform core** for grounded assistants in any industry.  
Product packs (HR, legal, support, etc.) are a **domain pack**: `config/` + `data/{tenant_id}/{domain_id}/`.

Canonical ops for local LLMs and caches: [LLM_PROVIDERS.md](./LLM_PROVIDERS.md).

---

## Layers

```
┌─────────────────────────────────────────────────────────┐
│  Platform core (this repo)                              │
│  Go orchestration · Python RAG · Redis · verify · CI    │
└───────────────────────────┬─────────────────────────────┘
                            │
              ┌─────────────┴─────────────┐
              ▼                           ▼
        Domain pack A              Domain pack B
        config + data/               config + data/
```

| Layer | Paths | Changes often? |
|-------|-------|----------------|
| **Core** | `server/`, `api/`, `rag/`, `migrations/`, `webapp/`, `scripts/` | No |
| **Domain pack** | `config/domains.json`, `config/locales/{ru,en}/`, `data/*` | **Yes** |
| **Optional runtime** | Redis (caches), Ollama / vLLM (local LLM), grounded-guardrails (`:50052`) | As needed |

**`domain_id`** — workspace / knowledge base identifier.  
**`tenant_id`** — multi-tenant isolation.

---

## Runtime services (Compose)

| Service | Role |
|---------|------|
| **server** (Go) | Auth, sessions, LLM call, verify (local or remote), citations, `/metrics` |
| **python** | HTTP RAG (`:5000`, Gunicorn) + **gRPC Retriever** (`:50051`) |
| **postgres** | Sessions, messages, analytics; optional pgvector |
| **redis** | Embedding cache + semantic LLM response cache |
| **webapp** | Reference UI (nginx → Go) |
| **ollama** / **vllm** | Optional local LLM (`--profile ollama` / `vllm`) |
| **guardrails** (optional) | [grounded-guardrails](https://github.com/kantik001/grounded-guardrails) gRPC `:50052` — `docker-compose.guardrails.yml` |

---

## Text chat flow

1. Client → Go `POST /message` (optional `?stream=1` for SSE)
2. On empty chat history: optional **semantic response cache** (Redis) → `X-Cache: HIT` and return
3. Else Go → Python `POST /rag/context` (`domain_id`, `tenant_id`, `locale`)
4. Python: embeddings (Redis-backed when `REDIS_URL` set) → vector store (Chroma / Qdrant / pgvector) → optional hybrid BM25+RRF / rerank → fragments
5. Go → OpenAI-compatible LLM (`LLM_PROVIDER` → OpenRouter / Ollama / vLLM)
6. **Verify** (after LLM):
   - `GUARDRAILS_MODE=local` (default): in-process Spec numeric check in `internal/rag/verify.go`
   - `remote` / `hybrid`: gRPC `VerifyText` → grounded-guardrails `:50052` (hybrid falls back to local on transport errors)
7. Disclaimer → Postgres (`citations[]`); verified answers may hit the response cache

**Agent path:** call gRPC `grounded.rag.v1.Retriever/Retrieve` on Python `:50051` (metadata `x-rag-service-token` when configured).

**Ports:** Retriever `:50051` · Guardrails (optional) `:50052` — see [GUARDRAILS.md](./GUARDRAILS.md) · [ECOSYSTEM.md](./ECOSYSTEM.md).

---

## Knowledge documents

Formats: **`.txt`**, **`.pdf`**, **`.docx`** → `rag/document_loaders.py` → chunking → configured vector backend.

**Production source of truth:** Postgres `kb_documents` + blob store (`KB_BLOB_DIR` or S3). See [KB_SOURCE_OF_TRUTH.md](./KB_SOURCE_OF_TRUTH.md).

**Legacy layout:** `data/{tenant_id}/{domain_id}/` — still dual-written on upload and used as ingest fallback.

**Indexes (Chroma, BM25):** disposable — rebuild via `POST /admin/ingest` without touching originals.

---

## New assistant from template pack

Prefer [packs/](../../packs/) over legacy `init_domain`:

```bash
python scripts/init_pack.py list
python scripts/init_pack.py install it_support   # or: hr, legal_faq
python scripts/reindex_rag.py
```

Registry: `packs/registry.yaml` — validate with `python scripts/init_pack.py registry --validate`.

---

## New domain checklist (manual)

1. Entry in `config/domains.json` (with `names.ru` / `names.en`)
2. Documents via admin upload (registry + blob) or `data/{tenant_id}/{domain_id}/` + `python scripts/backfill_kb_registry.py`
3. Locale bundles: `config/locales/ru/` and `config/locales/en/`
4. `POST /admin/ingest` or `python scripts/reindex_rag.py`
5. `eval/rag_{domain}_baseline.jsonl` + `make eval-retrieval`

Typical MVP estimate: **2–5 days** with documents ready.

---

## Documentation

- Providers / Redis / gRPC: [LLM_PROVIDERS.md](./LLM_PROVIDERS.md)
- Optional remote verify: [GUARDRAILS.md](./GUARDRAILS.md)
- Ecosystem map: [ECOSYSTEM.md](./ECOSYSTEM.md)
- Deploy: [DEPLOY.md](./DEPLOY.md)
- Network: [NETWORK_SECURITY.md](./NETWORK_SECURITY.md)
- English KB: [knowledge-base/README.md](./knowledge-base/README.md)
- Russian KB: [../ru/knowledge-base/README.md](../ru/knowledge-base/README.md)
