# Дорожная карта — Grounded LLM

## Сделано

### Ядро платформы
- Пайплайн документов: `.txt`, `.pdf`, `.docx`
- Админка: upload + reindex, citations в UI, eval baseline
- Удалён legacy API, учёт миграций в `schema_migrations`

### Фаза 1 — доверие к ответам
- Цитаты в чате, настройка `rag_k`, статистика индекса и удаление статей в админке, расширенный eval в CI

### Фаза 2 — интеграторы
- **SSE streaming** — `POST /message?stream=1` (Web App с fallback на JSON)
- **API keys** — `X-API-Key`, env `API_KEYS` или `API_KEYS_FILE`
- **API v1** — `/api/v1/*` + `GET /api/v1/openapi.json`
- **Мультитенантность (минимальная)** — `X-Tenant-ID`, `data/{tenant}/{domain}/`, фильтр Chroma по `tenant_id`
- **Наблюдаемость** — `X-Request-ID`, `GET /metrics` (Prometheus), структурированные логи
- **Админ: feedback** — `GET /admin/feedback`
- **Scaffold домена** — `scripts/init_domain.sh` / `init_domain.ps1`

### Локализация (i18n)
- Документация `docs/ru/` и `docs/en/`
- Конфиги промптов и UI: `config/locales/{ru,en}/`
- Middleware `X-Locale`, `Accept-Language`, `?locale=`

## Фаза 3 — платформа и монетизация (далее)

- Helm / Terraform, managed vector DB
- Open core vs hosted SaaS
- Vision domain pack, audit log, analytics dashboard

См. также: [ARCHITECTURE.md](./ARCHITECTURE.md), [DEPLOY.md](./DEPLOY.md).
