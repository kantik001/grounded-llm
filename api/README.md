# Python RAG service (`api/`)

Internal retrieval service for Grounded LLM (not the public Go `/api/v1` surface).

```
api/
├── http/app.py          # HTTP :5000 — /rag/context, /health, /ready, /metrics, admin
├── grpc/retriever.py    # gRPC :50051 — grounded.rag.v1.Retriever
├── gen/                 # generated stubs from proto/retriever.proto (do not edit)
├── proto/retriever.proto
├── auth.py              # shared service/admin secrets
├── retrieve_metrics.py  # Prometheus counters/histogram for retrieve
├── entrypoint.sh        # starts Gunicorn + gRPC under tini
└── requirements*.txt    # runtime + optional vector/connector extras
```

Engine code lives in `rag/`. Optional backends: `requirements-qdrant.txt`, `requirements-pgvector.txt`.

```bash
# HTTP only
python -m flask --app api.http.app run -p 5000

# gRPC only
python -m api.grpc

# both (same as Docker)
sh api/entrypoint.sh

# after editing proto/
pip install grpcio-tools==1.69.0
python scripts/gen_retriever_grpc.py
```

Guardrails contract (separate service `:50052`) lives at repo-root `proto/guardrails.proto`, not here.
