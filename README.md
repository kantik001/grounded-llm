# Grounded LLM

Universal **grounded LLM platform core**: answers grounded in your documents (RAG), sessions, Telegram Web App, admin upload, eval harness.

Not tied to any industry. The bundled `default` domain is a demo HR knowledge base in `data/default/`.

## Architecture

```
Telegram Web App  →  Go (auth, sessions, LLM orchestration, verify)
                         ↓
                    Python (RAG retrieval only)
                         ↓
                    Chroma + embeddings
                         ↓
                    data/{domain_id}/*.{txt,pdf,docx}
```

| Layer | Path | Purpose |
|-------|------|---------|
| **Core** | `server/`, `api/`, `rag/`, `migrations/`, `webapp/` | Orchestration, retrieval, reference UI |
| **Domain pack** | `config/`, `config/locales/{ru,en}/`, `data/{tenant}/{domain}/` | Prompts, branding, knowledge documents |

## Quick start

```bash
cp .env.example .env
# Set LLM_API_KEY, TELEGRAM_BOT_TOKEN (or TELEGRAM_AUTH_DISABLED=true for local dev)

docker compose up -d --build
python scripts/reindex_rag.py
```

| Service | URL |
|---------|-----|
| Web App | http://localhost/ |
| Go API | http://localhost:8080/health |
| Python | http://localhost:5000/health |

## API

- `GET /domains` — domain catalog
- `POST /session`, `GET /history`, `POST /message` — chat (`domain_id` in JSON)
- `GET /branding`, `GET /onboarding?domain_id=` — locale via `X-Locale`, `?locale=`, or `Accept-Language` (`ru` / `en`)
- Admin: `POST /admin/upload`, `POST /admin/reindex`

Integrators (Phase 2): `X-API-Key` + `X-Tenant-ID`, streaming `POST /message?stream=1`, OpenAPI at `/api/v1/openapi.json`, metrics at `/metrics`.

## New domain

1. Add entry to `config/domains.json`
2. Add `.txt`, `.pdf`, or `.docx` files under `data/{domain_id}/`
3. Update `config/locales/ru/` and `config/locales/en/` (prompts, few_shot, onboarding, branding)
4. Run `python scripts/reindex_rag.py`

## Development

```bash
cd server && go run .
python api/app.py
make test
make eval-retrieval
```

Documentation: [`docs/README.md`](docs/README.md) — [English](docs/en/) and [Русский](docs/ru/).

## Publish to GitHub

```powershell
# Install https://cli.github.com/ then:
gh auth login
powershell -ExecutionPolicy Bypass -File scripts/create_github_repo.ps1
```

Creates private repo **`grounded-llm`** on your account and pushes `main`.

## License

MIT — see [LICENSE](LICENSE).
