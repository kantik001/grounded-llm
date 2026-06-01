# `api/app.py` — Python RAG API

**Source:** `api/app.py`  
**Language:** Python (Flask)  
**Related:** `rag/retrieval.py`, `rag/vector_store.py`, `rag/domains_config.py`, `rag/document_loaders.py`  
**Called by:** Go server (`server/rag_chat.go`, `server/admin.go`)

---

## Purpose

Separate **Python HTTP service** (port **5000**, Compose service **`python`**).

| Endpoint | Purpose |
|----------|---------|
| `POST /rag/context` | Retrieval: article fragments for a question (no LLM) |
| `GET /domains` | Domain catalog from `config/domains.json` |
| `GET /health` | Healthcheck |
| `GET /admin/index-stats` | Chunks per file (`?domain_id=&tenant_id=`, `X-Admin-Secret`) |
| `POST /admin/reindex` | Rebuild Chroma (`X-Admin-Secret`) |

Go calls: `PYTHON_RAG_URL` → `http://python:5000/rag/context`.

**No CV/classify in core** — RAG retrieval only.

---

## `POST /rag/context`

JSON body:

```json
{
  "question": "...",
  "domain_id": "default",
  "tenant_id": "default",
  "locale": "en"
}
```

Response: `success`, `context`, `few_shot`, `fragments[]`, `category`, `error`.

Few-shot loaded from `config/locales/{locale}/few_shot.json`.

---

## `POST /admin/reindex`

Header `X-Admin-Secret` = env `ADMIN_SECRET`.

Chain: `reset_vector_store()` → `load_vector_store(force_reindex=True)`.

Indexes files from `data/{tenant_id}/{domain_id}/`: `.txt`, `.pdf`, `.docx`.

---

## Environment variables

| Variable | Purpose |
|----------|---------|
| `PYTHON_SERVICE_PORT` | port (default 5000) |
| `DOMAINS_CONFIG_PATH` | path to `domains.json` |
| `LOCALES_ROOT` | path to locale bundles |
| `DEFAULT_LOCALE` | default few-shot locale |
| `DEFAULT_TENANT_ID` | default tenant |
| `ADMIN_SECRET` | protect reindex |
| `FORCE_RAG_REINDEX` | full rebuild on startup |

---

## Run

```bash
# from repo root
python api/app.py
```

Docker: `CMD ["python", "api/app.py"]` in `Dockerfile.python`.

---

## What to read next

| Topic | File |
|-------|------|
| Indexing | [rag-vector_store.md](./rag-vector_store.md) |
| Search | [rag-retrieval.md](./rag-retrieval.md) |
| Domains | [rag-domains_config.md](./rag-domains_config.md) |
| Docker | [docker-overview.md](./docker-overview.md) |
