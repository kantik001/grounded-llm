# RAG engine (`rag/`)

Library used by the Python RAG **service** (`api/`) and by scripts (`reindex_rag.py`, eval). Not an HTTP/gRPC process by itself.

```
Postgres kb_documents + blobs  →  loaders → chunk → embed → vector (+ optional BM25)
                                              ↓
                              search / hybrid RRF / rerank → retrieve_rag_context
```

| Module | Role |
|--------|------|
| `document_loaders.py` | `.txt`, `.pdf`, `.docx` |
| `kb/documents.py` | Postgres registry, discover for ingest, manifest scan |
| `kb/index_runs.py` | Disposable index run pointers |
| `storage/blob_store.py` | Local or S3 blob backend |
| `indexing.py` | Load from registry + chunk (500/50) + stable `chunk_id` |
| `vector_backend/` | Chroma (default), Qdrant, pgvector |
| `vector_store.py` | Facade: index, search, hybrid, readiness |
| `sparse_index.py` | BM25 for `RAG_RETRIEVAL_MODE=hybrid` |
| `rrf.py` / `hybrid_rank.py` / `rerank.py` | Fusion and optional rerank |
| `embedding_cache.py` | Redis cache when `REDIS_URL` is set |
| `domains_config.py` | `config/domains.json` |
| `retrieval.py` | Context string + few-shot for callers |
| `ingest/` | Async pipeline: parse → staging → embed → index (`rag/ingest/`) |
| `verifier.py` | Local numeric check (tests / Spec path) |
| `kb_discovery.py` | **Migration only** — scan git-tracked `data/` for `backfill_kb_registry.py` |

**vs `api/`:** transport (HTTP `:5000`, gRPC `:50051`, auth, metrics) lives in [`api/`](../api/README.md). This package is the retrieval core.

## Environment (common)

| Variable | Default | Notes |
|----------|---------|--------|
| `VECTOR_STORE` | `chroma` | `chroma` \| `qdrant` \| `pgvector` |
| `CHROMA_PERSIST_DIR` | `./chroma_db` | Chroma persist path |
| `QDRANT_URL` / `QDRANT_COLLECTION` | `http://127.0.0.1:6333` / `grounded_llm` | Qdrant |
| `PGVECTOR_URL` or `DATABASE_URL` | — | pgvector DSN + KB registry |
| `PGVECTOR_COLLECTION` | `grounded_chunks` | Table/collection name |
| `RAG_RETRIEVAL_MODE` | `vector` | `vector` \| `hybrid` (dense + BM25 + RRF); Compose/CI default to `hybrid` |
| `RAG_RRF_K` | `60` | RRF constant |
| `RAG_RERANKER` | `none` | `none` \| `keyword` \| `cross_encoder`; Compose defaults to `keyword` |
| `RAG_CROSS_ENCODER_MODEL` | `cross-encoder/ms-marco-MiniLM-L-6-v2` | When cross-encoder on |
| `RAG_E5_PREFIXES` | auto | `query:`/`passage:` prefixes; auto-on for e5 models, index rebuilds on flip |
| `KB_BLOB_BACKEND` | `local` | `local` \| `s3` |
| `KB_BLOB_DIR` | `{DATA_DIR}/blobs` | Local blob root |
| `SPARSE_INDEX_DIR` | project `sparse_index/` | BM25 persist — see [`sparse_index/README.md`](../sparse_index/README.md) |
| `REDIS_URL` | — | Embedding cache |
| `EMBEDDING_CACHE_TTL_SEC` | `3600` | Cache TTL |
| `DOMAINS_CONFIG_PATH` | `config/domains.json` | Domain catalog |
| `LOCALES_ROOT` | `config/locales` | Few-shot JSON |
| `FORCE_RAG_REINDEX` | — | Force rebuild on load |
| `DEFAULT_TENANT_ID` | `default` | Tenant fallback |

Optional pip extras: `api/requirements-qdrant.txt`, `api/requirements-pgvector.txt`.

## Ops

```bash
python scripts/reindex_rag.py          # full rebuild from registry (dev/CI)
# POST /admin/ingest                  → async per-file pipeline (production)
# POST /admin/reindex                 → incremental sync job (dev fallback)
# POST /admin/reindex {"full":true}   → full rebuild
```

Ingestion: [docs/en/INGESTION.md](../docs/en/INGESTION.md).

Quality gates: [`eval/`](../eval/README.md). Architecture: [`docs/en/ARCHITECTURE.md`](../docs/en/ARCHITECTURE.md).
