# Project structure — Grounded LLM

High-level map of this repository. Deeper map (per-folder files): [docs/en/knowledge-base/PROJECT_STRUCTURE.md](docs/en/knowledge-base/PROJECT_STRUCTURE.md).

| Path | Purpose |
|------|---------|
| `server/` | Go: auth, sessions, RAG+LLM orchestration, admin, verify, optional guardrails client |
| `api/` | Python RAG **service** (internal): HTTP `:5000` + gRPC `:50051` — see `api/README.md` |
| `rag/` | Retrieval engine: vector backends, hybrid/BM25, rerank, loaders, domains config |
| `config/` | Runtime config: `domains.json`, locales, examples, schemas — see `config/README.md` |
| `data/{tenant}/{domain}/` | Knowledge base documents (`.txt`, `.pdf`, `.docx`) |
| `packs/` | Official template packs (HR, IT Support, Legal FAQ) + registry |
| `webapp/` | Reference UI: chat, admin, embed widget (nginx) |
| `migrations/` | PostgreSQL schema |
| `eval/` | Retrieval eval baselines (JSONL) + adversarial suites |
| `scripts/` | Reindex, eval runner, smoke/load/backup, pack CLI |
| `sdk/python/` | Python SDK + CLI (`grounded-llm`) |
| `connectors/` | Optional ingest: SharePoint, Google Drive, Confluence |
| `conformance/` | Spec / OpenAPI conformance CLI |
| `deploy/` | Helm chart + Terraform references |
| `docs/en/` | Primary documentation (architecture, deploy, knowledge base) |
| `docs/ru/` | Russian docs (legacy locale mirror) |

**Compose / images (repo root):** `docker-compose*.yml`, `Dockerfile.{server,python,webapp}`, `Makefile`.

Documentation index: [README.md](README.md) · [docs/en/ARCHITECTURE.md](docs/en/ARCHITECTURE.md) · [docs/en/knowledge-base/README.md](docs/en/knowledge-base/README.md) · [docs/en/ECOSYSTEM.md](docs/en/ECOSYSTEM.md)
