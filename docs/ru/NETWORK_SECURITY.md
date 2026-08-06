**Канон (EN):** [NETWORK_SECURITY.md](../en/NETWORK_SECURITY.md)

# Сетевая безопасность

Харденинг для production.

## Экспозиция сервисов

| Сервис | Порт | Публично? |
|--------|------|-----------|
| webapp (nginx) | 80/443 | Да — UI + `/api` proxy |
| Go server | 8080 | Предпочтительно только через nginx/ingress (`127.0.0.1:8080` в prod compose) |
| Python RAG (HTTP) | 5000 | **Нет** — internal / loopback (`docker-compose.prod.yml` снимает publish) |
| Python gRPC Retriever | 50051 | **Нет** — только internal |
| grounded-guardrails (опц.) | 50052 | **Нет** — internal; `docker-compose.guardrails.yml` |
| Redis | 6379 | **Нет** — internal (prod снимает publish; local — `127.0.0.1`) |
| Postgres | 5432 | **Нет** — только internal network |
| Ollama / vLLM | 11434 / 8000 | **Нет** — опциональные profiles, internal |

В Kubernetes — `NetworkPolicy`: server → python/postgres/redis (и опц. server → guardrails:50052; agents → python:50051).

**Production Compose:**

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build
```

`GROUNDED_ENV=production` — Go и Python **не стартуют** без `RAG_SERVICE_TOKEN`, `ADMIN_PASSWORD`, `ADMIN_SECRET`.

## Внутренняя аутентификация

При `RAG_SERVICE_TOKEN` (обязателен в prod):

- Go шлёт `X-RAG-Service-Token` на `/rag/context` и readiness
- Python отклоняет `/rag/context` и `/ready` без валидного токена
- gRPC Retriever ждёт тот же токен в metadata `x-rag-service-token` (или `Authorization: Bearer …`)
- `/health` остаётся без auth для liveness

Токен ≥32 байт — в secrets manager.

## nginx (webapp)

`webapp/nginx.conf`:

- `limit_req` на `/api/` (30 req/min per IP, burst 20)
- `X-Frame-Options: SAMEORIGIN`
- `X-Content-Type-Options: nosniff`
- CSP — scripts/styles same origin
- `Cache-Control: no-store` на HTML shell

## CORS

`CORS_ALLOWED_ORIGINS` — явные origins (**без** `*` в prod).

## TLS

Терминация на ingress/nginx. Пробрасывайте `X-Forwarded-Proto` для корректных OIDC redirect URL.

## Rate limiting

`RATE_LIMIT_REQUESTS_PER_MINUTE` — на authenticated user / API key.

## Admin

- `/admin.html` и `/api/admin/*` — ACL / VPN где возможно  
- OIDC SSO (`config/SSO.md`)  
- Ротация `ADMIN_SECRET` и `RAG_SERVICE_TOKEN` при компрометации  

## См. также

- [TRUST_CENTER.md (EN)](../en/TRUST_CENTER.md)
- [SECURITY_BRIEF.md](./SECURITY_BRIEF.md)
