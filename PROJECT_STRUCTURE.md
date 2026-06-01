# Структура проекта (Grounded LLM) / Project structure

| Путь / Path | Назначение / Purpose |
|-------------|----------------------|
| `server/` | Go: auth, sessions, RAG+LLM, admin, verify |
| `api/` | Python Flask: `/rag/context`, `/health`, reindex |
| `rag/` | Chroma, retrieval, `domains_config`, `document_loaders` |
| `config/` | Domain pack: `domains.json`, prompts, branding |
| `data/{domain_id}/` | KB: `.txt`, `.pdf`, `.docx` |
| `webapp/` | Telegram Web App (reference UI) |
| `migrations/` | PostgreSQL schema |
| `eval/` | RAG baseline tests |
| `scripts/` | reindex, eval runner, smoke |
| `docs/` | Architecture, deploy, knowledge-base |

Документация: [`README.md`](README.md), [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md), [`docs/knowledge-base/README.md`](docs/knowledge-base/README.md).
