# Protobuf contracts

gRPC IDL for services that live **outside** (or alongside) the Python RAG package layout.

| Contract | Port | Package | Stubs / clients |
|----------|------|---------|-----------------|
| [`guardrails.proto`](guardrails.proto) | `:50052` | `guardrails.v1` | Go: `server/gen/guardrails/v1/` |
| [`api/proto/retriever.proto`](../api/proto/retriever.proto) | `:50051` | `grounded.rag.v1` | Python: `api/gen/` |

Docs: [GUARDRAILS.md](../docs/en/GUARDRAILS.md) · [Python API (Retriever)](../docs/en/knowledge-base/python-api.md)

## Lint / breaking (Buf)

Workspace config: repo-root [`buf.yaml`](../buf.yaml) (includes `proto/` + `api/proto/`).

```bash
# Lint both modules (Docker; no local Buf install required)
docker run --rm -v "${PWD}:/workspace" -w /workspace bufbuild/buf:1.50.1 lint

# Breaking change vs main (PRs / local)
docker run --rm -v "${PWD}:/workspace" -w /workspace bufbuild/buf:1.50.1 \
  breaking --against '.git#branch=main'
```

CI runs the same checks in the `proto-lint` job.

## Regenerate stubs

**Retriever (Python):**

```bash
python scripts/gen_retriever_grpc.py
python scripts/gen_retriever_grpc.py --check   # CI
```

**Guardrails (Go)** — from `server/`, with `protoc-gen-go` and `protoc-gen-go-grpc` on `PATH`:

```bash
protoc \
  -I ../proto \
  --go_out=. --go_opt=module=grounded_llm_server \
  --go-grpc_out=. --go-grpc_opt=module=grounded_llm_server \
  ../proto/guardrails.proto
```

Output: `server/gen/guardrails/v1/*.pb.go` (commit generated files).
