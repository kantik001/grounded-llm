# Architecture: Grounded LLM

This repository is the **platform core** for grounded assistants in any industry.  
Product packs (HR, legal, support, etc.) are a **domain pack**: `config/` + `data/{tenant_id}/{domain_id}/`.

---

## Layers

```
┌─────────────────────────────────────────────────────────┐
│  Platform core (this repo)                              │
│  Go orchestration · Python RAG · verify · admin · CI    │
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
| **Optional** | Vision/CV (outside core) | As needed |

**`domain_id`** — workspace / knowledge base identifier.  
**`tenant_id`** — multi-tenant isolation (Phase 2).

---

## Text chat flow

1. Client → Go `POST /message` (optional `?stream=1` for SSE)
2. Go → Python `POST /rag/context` (`domain_id`, `tenant_id`, `locale`)
3. Chroma → fragments + locale-specific few-shot
4. Go → LLM → verify → disclaimer → Postgres (with `citations[]`)

---

## Knowledge documents

Formats: **`.txt`**, **`.pdf`**, **`.docx`** → `rag/document_loaders.py` → chunking → Chroma.

Layout: `data/{tenant_id}/{domain_id}/` (legacy `data/{domain_id}/` still supported).

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
2. Documents in `data/{tenant_id}/{domain_id}/`
3. Locale bundles: `config/locales/ru/` and `config/locales/en/`
4. `python scripts/reindex_rag.py` or `scripts/init_domain.ps1`
5. `eval/rag_{domain}_baseline.jsonl` + `make eval-retrieval`

Typical MVP estimate: **2–5 days** with documents ready.

---

## Documentation

English: [knowledge-base/README.md](./knowledge-base/README.md).  
Russian: [../ru/knowledge-base/README.md](../ru/knowledge-base/README.md).
