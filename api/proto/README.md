# Retriever protobuf (Python RAG gRPC)

Source of truth for `grounded.rag.v1.Retriever` (`:50051`).

```bash
python scripts/gen_retriever_grpc.py
```

Output: `api/gen/retriever_pb2.py`, `api/gen/retriever_pb2_grpc.py`.

Guardrails (`:50052`) lives at repo-root [`proto/guardrails.proto`](../../proto/guardrails.proto).
Map + Buf lint: [`proto/README.md`](../../proto/README.md), root [`buf.yaml`](../../buf.yaml).
