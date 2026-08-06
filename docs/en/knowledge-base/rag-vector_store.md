# `rag/vector_store.py` — vector store

**Sources:** `rag/vector_store.py`, `rag/vector_backend/` (Chroma / Qdrant / pgvector), `rag/indexing.py`, `rag/sparse_index.py`  
**Data:** `data/{tenant_id}/{domain_id}/*.{txt,pdf,docx}`  
**Ops guide (env):** [../VECTOR_STORE.md](../VECTOR_STORE.md)  
**Called by:** `rag/retrieval.py`, admin reindex, `scripts/reindex_rag.py`

---

## Purpose

Core **vector RAG**: documents → embeddings → selected backend. No LLM here.  
Backend via `VECTOR_STORE=chroma|qdrant|pgvector` (default Chroma).

---

## Indexing pipeline

```mermaid
flowchart LR
    A[data/tenant/domain/*] --> B[document_loaders]
    B --> C[metadata tenant_id domain_id filename]
    C --> D[RecursiveCharacterTextSplitter]
    D --> E[chunk 500 overlap 50]
    E --> F[HuggingFaceEmbeddings e5-small]
    F --> G[Backend: Chroma / Qdrant / pgvector]
```

### `rag/document_loaders.py`

| Extension | Loader |
|-----------|--------|
| `.txt` | `TextLoader` (UTF-8) |
| `.pdf` | `PyPDFLoader` |
| `.docx` | `Docx2txtLoader` |

Metadata: `filename`, `domain_id`, `tenant_id`, `source_file`, `file_type`.

---

## Backends (`rag/vector_backend/`)

| Backend | Persist | When to use |
|---------|---------|-------------|
| **Chroma** | `chroma_db/` (volume `chroma_data`) | Compose / CI / local |
| **Qdrant** | managed Qdrant (`QDRANT_URL`) | separate vector DB |
| **pgvector** | Postgres (`009_pgvector.sql`) | one Postgres for sessions + vectors |

Changing backend or embedding model requires **full reindex** and eval re-run.

```bash
VECTOR_STORE=chroma python scripts/reindex_rag.py
pip install -r api/requirements-qdrant.txt
VECTOR_STORE=qdrant QDRANT_URL=http://127.0.0.1:6333 FORCE_RAG_REINDEX=true python scripts/reindex_rag.py
pip install -r api/requirements-pgvector.txt
VECTOR_STORE=pgvector FORCE_RAG_REINDEX=true python scripts/reindex_rag.py
```

---

## Hybrid retrieval (BM25 + dense + RRF)

When `RAG_RETRIEVAL_MODE=hybrid`:

1. dense hits from vector backend  
2. BM25 hits from `sparse_index/` (`SPARSE_INDEX_DIR`)  
3. RRF fusion (`RAG_RRF_K`, default 60)  
4. optional rerank: `RAG_RERANKER=keyword|cross_encoder`

Details: [../VECTOR_STORE.md](../VECTOR_STORE.md).

---

## `load_vector_store(force_reindex=False)`

| Situation | Behavior |
|-----------|----------|
| RAM cache | return cached |
| `FORCE_RAG_REINDEX=true` | rebuild backend index |
| else | open existing / create |

---

## `search(query, domain_id, tenant_id, k=8)`

Metadata filter is always `domain_id` + `tenant_id`.

---

## Docker

- `./data:/app/data:ro` (python)
- `chroma_data:/app/chroma_db` (when Chroma)
- `./data:/app/data` rw (server) — admin upload
- After upload — **reindex required**

---

## Dependencies

| Backend | File |
|---------|------|
| Chroma (default) | `api/requirements.txt` |
| Qdrant | `api/requirements-qdrant.txt` |
| pgvector | `api/requirements-pgvector.txt` |

---

## What to read next

| Topic | File |
|-------|------|
| Env / hybrid / rerank | [../VECTOR_STORE.md](../VECTOR_STORE.md) |
| Domains | [rag-domains_config.md](./rag-domains_config.md) |
| Retrieval | [rag-retrieval.md](./rag-retrieval.md) |
| HTTP reindex | [python-api.md](./python-api.md) |
