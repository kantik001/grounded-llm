# Vector store adapter

Grounded LLM supports pluggable vector indexes for the Python RAG service. The reference implementation uses **Chroma** (local persist). **Qdrant** is available as an optional backend for teams that operate a managed vector database.

---

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `VECTOR_STORE` | `chroma` | Backend: `chroma` or `qdrant` |
| `CHROMA_PERSIST_DIR` | `./chroma_db` | Chroma on-disk path |
| `QDRANT_URL` | `http://127.0.0.1:6333` | Qdrant HTTP endpoint |
| `QDRANT_COLLECTION` | `grounded_llm` | Qdrant collection name |
| `RAG_RETRIEVAL_MODE` | `vector` | `vector` or `hybrid` (keyword rerank) |
| `FORCE_RAG_REINDEX` | `false` | Rebuild index on startup |

---

## Chroma (default)

Used in Docker Compose, Helm, and CI (`eval-retrieval-gate`).

```bash
VECTOR_STORE=chroma python scripts/reindex_rag.py
```

Metadata filter: `domain_id` + `tenant_id` on every chunk.

---

## Qdrant (optional)

Install optional dependencies:

```bash
pip install -r api/requirements-qdrant.txt
```

Run Qdrant (example):

```bash
docker run -p 6333:6333 qdrant/qdrant
```

Reindex:

```bash
VECTOR_STORE=qdrant QDRANT_URL=http://127.0.0.1:6333 FORCE_RAG_REINDEX=true python scripts/reindex_rag.py
```

Changing backend or embedding model requires **full reindex** and eval gate re-run.

---

## Hybrid retrieval

When `RAG_RETRIEVAL_MODE=hybrid`, the service:

1. Fetches `2× rag_k` vector hits  
2. Reranks by keyword overlap with the query  
3. Returns top `rag_k` fragments  

No extra ML model — useful for policy docs with distinctive numbers and terms. Measure impact with `scripts/run_rag_eval.py` before enabling in production.

---

## Implementation

| Module | Role |
|--------|------|
| `rag/vector_backend/` | Backend interface + Chroma/Qdrant |
| `rag/vector_store.py` | Public API (`search`, `index_stats_for_domain`) |
| `rag/hybrid_rank.py` | Keyword reranking |

---

## Related

- [COMPATIBILITY.md](./COMPATIBILITY.md)
- [BACKUP_RESTORE.md](./BACKUP_RESTORE.md) — include Chroma PVC or Qdrant snapshots
