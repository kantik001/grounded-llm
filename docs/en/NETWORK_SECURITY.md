# Network security

Hardening guidance for production deployments.

## Service exposure

| Service | Port | Expose publicly? |
|---------|------|------------------|
| webapp (nginx) | 80/443 | Yes — UI + `/api` proxy |
| Go server | 8080 | Prefer via nginx/ingress only (`127.0.0.1:8080` in prod compose) |
| Python RAG | 5000 | **No** — internal network only (`docker-compose.prod.yml` removes host publish; local compose binds `127.0.0.1` only) |
| Postgres | 5432 | **No** — internal network only |

In Kubernetes, use `NetworkPolicy` to allow server → python/postgres only.

**Production Compose:**

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build
```

Set `GROUNDED_ENV=production` so Go and Python **refuse to start** without `RAG_SERVICE_TOKEN`, `ADMIN_PASSWORD`, and `ADMIN_SECRET`.

## Internal authentication

When `RAG_SERVICE_TOKEN` is set (required in production):

- Go sends `X-RAG-Service-Token` on `/rag/context` and readiness probes
- Python rejects `/rag/context` and `/ready` without a valid token
- `/health` stays unauthenticated for container liveness checks

Generate a strong random token (≥32 bytes) and store in your secrets manager.

## nginx (webapp)

The bundled `webapp/nginx.conf` sets:

- `limit_req` on `/api/` (30 req/min per IP, burst 20)
- `X-Frame-Options: SAMEORIGIN`
- `X-Content-Type-Options: nosniff`
- `Content-Security-Policy` — restricts scripts/styles to same origin
- `Cache-Control: no-store` on HTML shell

Adjust CSP if you embed third-party analytics or fonts.

## CORS

Configure `CORS_ALLOWED_ORIGINS` to explicit origins (no `*` in production). The Go server validates origins on API routes.

## TLS termination

Terminate TLS at ingress or nginx. Forward `X-Forwarded-Proto` so OIDC redirect URLs remain correct.

## Rate limiting

`RATE_LIMIT_REQUESTS_PER_MINUTE` applies per authenticated user/API key on protected routes.

## Admin surface

- Protect `/admin.html` and `/api/admin/*` with network ACLs or VPN where possible
- Enable OIDC for SSO (`config/SSO.md`)
- Rotate `ADMIN_SECRET` and `RAG_SERVICE_TOKEN` on compromise

## Related

- [TRUST_CENTER.md](./TRUST_CENTER.md)
- [SECURITY_BRIEF.md](./SECURITY_BRIEF.md)
