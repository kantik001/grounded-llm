# Структура проекта (Grounded LLM)

| Путь | Назначение |
|------|------------|
| `server/` | Go: auth, sessions, RAG+LLM, admin, verify |
| `api/` | Python Flask: `/rag/context`, `/health`, reindex |
| `rag/` | Chroma, retrieval, domains config |
| `config/` | Domain pack: `domains.json`, prompts, branding |
| `data/{domain_id}/` | Текстовые документы базы знаний |
| `webapp/` | Telegram Web App (reference UI) |
| `migrations/` | PostgreSQL schema |
| `eval/` | RAG baseline tests |
| `scripts/` | reindex, eval runner |

Документация: [`README.md`](README.md), [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md).
