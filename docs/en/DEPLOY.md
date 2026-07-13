# Deployment

Guide for running a **new project** on the Grounded LLM scaffold.  
Architecture: [ARCHITECTURE.md](./ARCHITECTURE.md).

---

## Quick start (Docker)

```bash
cp .env.example .env
# LLM_API_KEY, TELEGRAM_BOT_TOKEN (or TELEGRAM_AUTH_DISABLED=true for dev)

docker compose up -d --build
```

| Service | URL |
|---------|-----|
| Web App | http://localhost/ |
| Go API | http://localhost:8080/health |
| Python RAG | http://localhost:5000/health |

After adding documents under `data/`:

```bash
python scripts/reindex_rag.py
# or POST /admin/reindex (Basic auth + ADMIN_SECRET for Python)
```

Supported KB formats: **`.txt`**, **`.pdf`**, **`.docx`**.

---

## Config without rebuild

The `./config` directory is mounted into containers as `/config` (read-only).

| Variable | File / path |
|----------|-------------|
| `DOMAINS_CONFIG_PATH` | `domains.json` |
| `LOCALES_ROOT` | `config/locales` (`ru/`, `en/`) |
| `DEFAULT_LOCALE` | `ru` or `en` |
| `DEFAULT_TENANT_ID` | default tenant for KB paths |
| `API_KEYS` or `API_KEYS_FILE` | integrator API keys (Phase 2) |

**Reload Go without restart:**

```bash
docker compose kill -s HUP server
```

Or set `CONFIG_RELOAD_INTERVAL_SEC=300` in `.env`.

Python `rag/domains_config.py` reloads `domains.json` when mtime changes.

---

## Local development (without Docker)

1. Postgres + `.env` with `DATABASE_URL`.
2. `cd server && go run .`
3. Python: `python api/app.py` (from repo root).
4. Web: nginx or `webapp/` + `TELEGRAM_AUTH_DISABLED=true`, API on `:8080`.

---

## Eval after KB changes

```bash
pip install requests
set PYTHON_RAG_URL=http://localhost:5000/rag/context
python scripts/run_rag_eval.py --suite default
make eval-retrieval
```

Results: `eval/results/YYYYMMDD_HHMMSS.json`.

Run after: reindex, locale prompt changes, `LLM_MODEL` change.

---

## New customer: domain pack

### 1. Repository

```bash
git clone <url> client-assistant
cd client-assistant
```

### 2. Domain pack

| Action | Path |
|--------|------|
| KB documents | `data/{tenant_id}/{domain_id}/` (`.txt`, `.pdf`, `.docx`) |
| Domain catalog | `config/domains.json` |
| Prompts & few-shot | `config/locales/ru/`, `config/locales/en/` |
| UI branding | locale `branding.json`; customize `webapp/` if needed |
| Eval questions | `eval/rag_{domain}_baseline.jsonl` |

Scaffold: `python scripts/init_pack.py install <pack_id>` (preferred) or `scripts/init_domain.ps1` / `init_domain.sh`.

### 3. Index and verify

```bash
python scripts/reindex_rag.py
python scripts/run_rag_eval.py --suite default
```

### 4. Secrets

`.env`: `LLM_API_KEY`, `DATABASE_URL`, `CORS`, Telegram, `ADMIN_PASSWORD`, `ADMIN_SECRET`, optional `API_KEYS`.

### 5. Pilot metrics

Verify pass rate, “not in materials” rate, thumbs up/down, latency p95.  
Prometheus: `GET /metrics`.

---

## Smoke

```bash
make smoke
# TELEGRAM_AUTH_DISABLED=true, server on :8080
```

---

## Do not copy to a new instance

- volume `chroma_data` (recreated by reindex).
- `postgres_data` / production sessions.
- `.env` secrets — only `.env.example` as template.

---

## Optional modules

**Vision / CV** — separate domain pack, not part of platform core.

**Hosted SaaS signup** — disabled by default. To enable self-serve tenant creation + Stripe billing, see [SAAS.md](./SAAS.md) and [BILLING.md](./BILLING.md). Not required for on-prem pilots.
