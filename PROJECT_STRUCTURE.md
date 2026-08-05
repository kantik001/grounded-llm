# Project structure — Grounded LLM

High-level map of this repository. Deeper map (per-folder files): [docs/en/knowledge-base/PROJECT_STRUCTURE.md](docs/en/knowledge-base/PROJECT_STRUCTURE.md).

| Path | Purpose |
|------|---------|
| `server/` | Go API orchestrator (`cmd/server` + `internal/`) — see `server/README.md` |
| `api/` | Python RAG **service** (internal): HTTP `:5000` + gRPC `:50051` — see `api/README.md` |
| `proto/` | Guardrails gRPC IDL (`:50052`); Retriever IDL is `api/proto/` — see `proto/README.md`, `buf.yaml` |
| `rag/` | Retrieval **engine** (library): backends, hybrid/BM25, rerank — service is `api/` — see `rag/README.md` |
| `config/` | Runtime config: `domains.json`, locales, examples, schemas — see `config/README.md` |
| `data/{tenant}/{domain}/` | Knowledge base documents — see `data/README.md` |
| `packs/` | Official template packs + registry — see `packs/README.md` |
| `webapp/` | Reference UI: chat, admin, embed (nginx) — see `webapp/README.md` |
| `site/` | GitHub Pages landing — see `site/README.md` |
| `migrations/` | PostgreSQL schema — see `migrations/README.md` |
| `eval/` | Retrieval eval baselines (JSONL) + adversarial — see `eval/README.md` |
| `tests/` | Python pytest (unit/integration) — see `tests/README.md` |
| `scripts/` | Reindex, eval, smoke/load, pack CLI, CI helpers — see `scripts/README.md` |
| `sdk/python/` | Python SDK + CLI + `examples/` — see `sdk/README.md` |
| `connectors/` | Optional ingest (SharePoint, Drive, Confluence) — see `connectors/README.md` |
| `conformance/` | Spec / OpenAPI conformance CLI — see `conformance/README.md` |
| `deploy/` | Helm chart + Terraform references — see `deploy/README.md` |
| `models/` | Optional local weight mount (gitignored binaries) — see `models/README.md` |
| `sparse_index/` | Runtime BM25 cache (gitignored `.pkl`) — see `sparse_index/README.md` |
| `.github/` | CI, Dependabot, CODEOWNERS, issue/PR templates — see `.github/AUTOMATION.md` |
| `docs/en/` | Primary documentation (architecture, deploy, knowledge base) |
| `docs/ru/` | Russian docs (legacy locale mirror) |

**Compose / images (repo root):** `docker-compose*.yml`, `Dockerfile.{server,python,webapp}`, `Makefile`.

Documentation index: [README.md](README.md) · [docs/en/ARCHITECTURE.md](docs/en/ARCHITECTURE.md) · [docs/en/knowledge-base/README.md](docs/en/knowledge-base/README.md) · [docs/en/ECOSYSTEM.md](docs/en/ECOSYSTEM.md)
