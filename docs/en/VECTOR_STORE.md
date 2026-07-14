# Vector store adapter

Grounded LLM supports pluggable vector indexes for the Python RAG service. The reference implementation uses **Chroma** (local persist). **Qdrant** is available as an optional backend for teams that operate a managed vector database.

---

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `VECTOR_STORE` | `chroma` | Backend: `chroma`, `qdrant`, or `pgvector` |
| `CHROMA_PERSIST_DIR` | `./chroma_db` | Chroma on-disk path |
| `QDRANT_URL` | `http://127.0.0.1:6333` | Qdrant HTTP endpoint |
| `QDRANT_COLLECTION` | `grounded_llm` | Qdrant collection name |
| `PGVECTOR_URL` | `DATABASE_URL` | Postgres DSN for pgvector (`postgresql+psycopg://...`) |
| `PGVECTOR_COLLECTION` | `grounded_chunks` | pgvector collection name |
| `RAG_RETRIEVAL_MODE` | `vector` | `vector` or `hybrid` (BM25 + dense + RRF) |
| `RAG_RRF_K` | `60` | RRF constant for hybrid fusion |
| `SPARSE_INDEX_DIR` | `./sparse_index` | BM25 index persistence path |
| `RAG_RERANKER` | `none` | `none`, `keyword`, or `cross_encoder` (optional second stage) |
| `RAG_CROSS_ENCODER_MODEL` | `cross-encoder/ms-marco-MiniLM-L-6-v2` | Cross-encoder model name |
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

## pgvector (optional)

Store embeddings in PostgreSQL with the **pgvector** extension — same database as sessions when desired.

**Requirements:**

- Postgres image with pgvector (reference deploy uses `pgvector/pgvector:pg16`)
- Migration `009_pgvector.sql` enables `CREATE EXTENSION vector`

Install optional dependencies:

```bash
pip install -r api/requirements-pgvector.txt
```

Reindex:

```bash
VECTOR_STORE=pgvector \
DATABASE_URL=postgres://grounded:grounded@localhost:5432/grounded?sslmode=disable \
FORCE_RAG_REINDEX=true \
python scripts/reindex_rag.py
```

Docker Compose (optional):

```bash
VECTOR_STORE=pgvector docker compose up -d --build
```

Chunks use stable `chunk_id` metadata; filters: `domain_id` + `tenant_id`.

---

## Reranking and hybrid retrieval

| Mode | Env | Notes |
|------|-----|-------|
| Vector only | `RAG_RETRIEVAL_MODE=vector` (default) | Top-k from embedding search |
| **Hybrid (BM25 + dense + RRF)** | `RAG_RETRIEVAL_MODE=hybrid` | Sparse BM25 + dense vectors fused with reciprocal rank fusion |
| Keyword rerank (optional) | `RAG_RERANKER=keyword` | Second-stage overlap rerank after vector or hybrid fusion |
| Cross-encoder | `RAG_RERANKER=cross_encoder` | `sentence-transformers` CrossEncoder; slower, often better on policy Q&A |

### Hybrid flow (`RAG_RETRIEVAL_MODE=hybrid`)

1. Fetch `3× rag_k` hits from dense vector search (Chroma/Qdrant)  
2. Fetch `3× rag_k` hits from BM25 sparse index (`rag/sparse_index.py`)  
3. Fuse with RRF (`RAG_RRF_K`, default `60`)  
4. Optionally rerank with `RAG_RERANKER`  
5. Return top `rag_k` fragments  

Sparse index is rebuilt with `FORCE_RAG_REINDEX=true` or `python scripts/reindex_rag.py` and persisted under `sparse_index/` (override with `SPARSE_INDEX_DIR`).

Measure impact:

```bash
RAG_RETRIEVAL_MODE=hybrid python scripts/run_rag_eval.py --suite hybrid
RAG_RETRIEVAL_MODE=hybrid python scripts/run_rag_eval.py --suite default_en
```

CI uses default (`vector`) so existing gates stay fast.

---

## Implementation

| Module | Role |
|--------|------|
| `rag/indexing.py` | Shared chunking + `chunk_id` metadata |
| `rag/sparse_index.py` | BM25 sparse index |
| `rag/rrf.py` | Reciprocal rank fusion |
| `rag/vector_backend/` | Backend interface + Chroma / Qdrant / pgvector |
| `rag/vector_store.py` | Public API (`search`, `index_stats_for_domain`) |
| `rag/rerank.py` | Keyword + cross-encoder reranking |
| `rag/hybrid_rank.py` | Back-compat keyword rerank re-export |

---

## Related

- [COMPATIBILITY.md](./COMPATIBILITY.md)
- [BACKUP_RESTORE.md](./BACKUP_RESTORE.md) — include Chroma PVC or Qdrant snapshots
