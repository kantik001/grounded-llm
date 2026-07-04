# Project structure — Grounded LLM

| Path | Purpose |
|------|---------|
| `server/` | Go: auth, sessions, RAG+LLM orchestration, admin, verify |
| `api/` | Python Flask: `/rag/context`, `/health`, reindex |
| `rag/` | Chroma, retrieval, `domains_config`, `document_loaders` |
| `config/` | Domain pack: `domains.json`, prompts, branding, RBAC/SSO |
| `data/{tenant}/{domain}/` | Knowledge base: `.txt`, `.pdf`, `.docx` |
| `packs/` | Official template packs (HR, IT Support) |
| `webapp/` | Reference Telegram Web App UI |
| `migrations/` | PostgreSQL schema |
| `eval/` | RAG baseline eval suites (JSONL) |
| `scripts/` | Reindex, eval runner, smoke tests, pack CLI |
| `docs/en/` | Primary documentation (architecture, deploy, knowledge base) |
| `docs/ru/` | Russian docs (legacy locale mirror) |

Documentation index: [README.md](README.md) · [docs/en/ARCHITECTURE.md](docs/en/ARCHITECTURE.md) · [docs/en/knowledge-base/README.md](docs/en/knowledge-base/README.md)
