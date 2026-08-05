# `tests/` — Python unit / integration tests

Pytest suite for `rag/`, `api/`, packs, connectors, schemas. **Not** retrieval quality baselines — those live in [`eval/`](../eval/README.md).

| Area | Examples |
|------|----------|
| RAG engine | `test_rrf`, `test_sparse_index`, `test_hybrid_*`, `test_indexing`, `test_rerank` |
| Vector backends | `test_vector_*`, `test_pgvector_backend` |
| API / gRPC | `test_api_ready`, `test_grpc_retriever`, `test_retrieve_metrics` |
| Config / packs | `test_domains_*`, `test_pack_registry`, `test_init_pack`, `test_plans_yaml` |
| Connectors | `test_*connector*`, `test_export_connectors` |
| Eval plumbing | `test_eval_baseline` (JSON schema), `test_rag_eval_check` |
| Conformance | `test_conformance_cli` |

Go tests: `server/internal/...` (`cd server && go test ./...`). SDK: `sdk/python/tests/`.

```bash
pip install -r tests/requirements-test.txt
pytest tests/ -v --tb=short
# or
make test-py
```

`conftest.py` adds the repo root to `PYTHONPATH`. More detail: [tests-overview.md](../docs/en/knowledge-base/tests-overview.md).
