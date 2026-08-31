# `webapp/` — reference UI

Static front-end served by nginx (`Dockerfile.webapp`, Compose service `webapp` on `:80`). Talks to the **Go API** via `/api/` proxy — not to Python `api/` directly.

| File | Role |
|------|------|
| `index.html` + `app.js` / `app.css` | Chat (Telegram Web App / browser) |
| `admin.html` + `admin.js` | KB upload, ingest, reindex, admin ops |
| `signup.html` + `signup.js` / `signup.css` | SaaS signup flow |
| `embed.html` + `embed.js` / `embed.css` | Embeddable widget |
| `nginx.conf` | Static root + `/api/` → `server:8080` |

Deep dive: [webapp-overview.md](../docs/en/knowledge-base/webapp-overview.md) · Docker: [docker-overview.md](../docs/en/knowledge-base/docker-overview.md).

```bash
# Via Compose (typical)
docker compose up -d webapp
# Open http://localhost (or mapped host port)
```
