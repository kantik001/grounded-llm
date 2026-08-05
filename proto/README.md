# Protobuf contracts outside the Python RAG package

| File | Owner | Notes |
|------|-------|--------|
| `guardrails.proto` | Guardrails gRPC service (`:50052`), Go client under `server/gen/guardrails/` | Not part of Python Retriever |
| `api/proto/retriever.proto` | Python RAG Retriever (`:50051`) | Generate stubs into `api/gen/` |

Retriever stubs: `python scripts/gen_retriever_grpc.py`
