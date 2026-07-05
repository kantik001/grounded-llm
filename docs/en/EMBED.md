# Embeddable chat widget

Lightweight **iframe-friendly** chat UI for intranet portals — uses the same `/api/session` and `/api/message` endpoints as the main webapp.

---

## Quick start

1. Deploy Grounded LLM (Docker Compose or K8s).
2. Allow your portal origin in `CORS_ALLOWED_ORIGINS` on the Go server.
3. Embed:

```html
<iframe
  src="https://your-host/embed.html?api=/api/&tenant=default&locale=en"
  width="420"
  height="560"
  style="border:0;border-radius:12px"
  title="Grounded assistant"
></iframe>
```

Local dev: `http://localhost/embed.html?api=/api/`

---

## Query parameters

| Param | Default | Description |
|-------|---------|-------------|
| `api` | `/api/` | API base path (must proxy to Go server) |
| `tenant` | `default` | `X-Tenant-ID` |
| `locale` | `en` | `X-Locale` |

---

## Security notes

- `embed.html` uses a relaxed `frame-ancestors *` CSP — **tune for production** (restrict to your intranet domains).
- Enable Telegram auth or API keys as required by your deployment (`TELEGRAM_AUTH_DISABLED`, `API_KEYS`).
- Do not expose admin routes through the same origin without edge ACLs.

---

## Files

| File | Role |
|------|------|
| `webapp/embed.html` | Shell |
| `webapp/embed.css` | Compact layout |
| `webapp/embed.js` | Session + message flow |

Nginx: see `webapp/nginx.conf` location `=/embed.html`.

---

## Related

- [DEPLOY.md](./DEPLOY.md)
- [NETWORK_SECURITY.md](./NETWORK_SECURITY.md)
