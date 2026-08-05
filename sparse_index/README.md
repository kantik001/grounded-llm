# Local BM25 sparse index (runtime)

On-disk cache for hybrid retrieval (`RAG_RETRIEVAL_MODE=hybrid`). Built by the Python RAG stack from `data/` — **do not commit** `.pkl` payloads.

| Item | Notes |
|------|--------|
| Default path | `./sparse_index/bm25_index.pkl` (repo root) |
| Override | `SPARSE_INDEX_DIR` (see [`rag/README.md`](../rag/README.md)) |
| Rebuild | `python scripts/reindex_rag.py` or admin reindex |

Tracked in git: only this README (and `.gitkeep`). Ignored: everything else under `sparse_index/` (see `.gitignore`).
