# Миграции PostgreSQL (`migrations/*.sql`)

**Папка:** `migrations/` — также [`migrations/README.md`](../../../migrations/README.md)  
**Файлы:** `001`–`003`, `005`–`011` (нет `004`; см. [EN migrations-overview](../en/knowledge-base/migrations-overview.md))  
**Кто применяет:** Go-сервер при старте (`internal/app/main.go` → `store.RunAllMigrations`)  
**СУБД:** PostgreSQL 16 (контейнер `postgres` в `docker-compose.yml`)

---

## Что такое миграция простыми словами

**Миграция** — SQL-скрипт, который **меняет структуру базы данных** (таблицы, колонки, индексы).

Зачем не править БД руками в pgAdmin:

- одна и та же схема у вас, у коллеги и на сервере;
- изменения в Git — видно историю;
- при новом деплое сервер сам накатывает скрипты.

Файлы нумеруются по порядку — это эволюция схемы, а не несколько разных баз.

---

## Как миграции запускаются в этом проекте

```mermaid
sequenceDiagram
    participant DC as docker compose up server
    participant Go as server (main.go + postgres_store.go)
    participant PG as PostgreSQL

    DC->>Go: старт контейнера
    Go->>PG: подключение DATABASE_URL
    Go->>Go: findMigrationsDir → /migrations
    Go->>Go: sort 001, 002, 003
    loop каждый .sql
        Go->>PG: выполнить весь файл целиком
    end
    Go->>Go: API готов
```

### Важные детали

1. Таблица **`schema_migrations`** — каждый `.sql` применяется **один раз**.
2. В скриптах — **`IF NOT EXISTS`** / **`ADD COLUMN IF NOT EXISTS`** на случай ручного восстановления.

### Где лежат файлы в Docker

- `Dockerfile.server`: `COPY migrations /migrations`
- `docker-compose.yml`: `MIGRATIONS_DIR=/migrations`

Локально без Docker Go ищет папку `migrations` или `../migrations`.

---

## Файл `001_init.sql` — фундамент

Три таблицы + связи.

### `users` — кто пишет в чат

| Колонка | Назначение |
|---------|------------|
| `id` | внутренний id в БД |
| `telegram_id` | id из Telegram, **UNIQUE** |
| `username`, `first_name`, `last_name` | профиль |
| `created_at`, `updated_at` | метки времени |

### `chat_sessions` — один «диалог»

| Колонка | Назначение |
|---------|------------|
| `id` | TEXT (случайный hex из Go) |
| `user_id` | → `users.id`, CASCADE при удалении user |
| `created_at`, `updated_at` | когда открыли/обновили сессию |

Индекс `idx_chat_sessions_user_id` — быстро найти все сессии пользователя.

### `messages` — сообщения в сессии

| Колонка | Назначение |
|---------|------------|
| `id` | BIGSERIAL |
| `session_id` | → `chat_sessions.id` |
| `role` | `user` или `assistant` |
| `content` | текст |
| `kind` | тип сообщения |
| `image_token` | ссылка на файл (domain pack / vision) |
| `class_prediction`, `class_confidence` | опционально для vision domain pack |
| `created_at` | порядок в чате |

Индекс `(session_id, created_at)` — история чата по времени.

---

## Файл `002_domain_id.sql` — домен сессии

Добавляет колонку `domain_id` в `chat_sessions`:

```sql
ALTER TABLE chat_sessions
    ADD COLUMN IF NOT EXISTS domain_id TEXT NOT NULL DEFAULT 'default';

CREATE INDEX IF NOT EXISTS idx_chat_sessions_domain_id ON chat_sessions (domain_id);
```

Каждая сессия привязана к knowledge domain из `config/domains.json`.

---

## Файл `003_feedback_analytics.sql` — UX и метрики

### `message_feedback` — 👍 / 👎

| Колонка | Назначение |
|---------|------------|
| `message_id` | → `messages.id`, CASCADE |
| `user_id` | → `users.id` |
| `rating` | `-1` или `1` |
| `UNIQUE (message_id, user_id)` | один голос пользователя на сообщение |

### `analytics_events` — события для статистики

| Колонка | Назначение |
|---------|------------|
| `event_type` | строка-код события |
| `payload` | JSONB |
| `user_id` | опционально, SET NULL если user удалён |

---

## Порядок файлов

```
001_init.sql
002_domain_id.sql
003_feedback_analytics.sql
005_message_citations.sql
006_tenant_id.sql
```

Go сортирует по имени. **Новая миграция:** например `007_что_то.sql` — не менять старые файлы после деплоя в прод.

---

## `005_message_citations.sql`

Колонка `citations JSONB` у сообщений ассистента — фрагменты KB для UI.

---

## `006_tenant_id.sql`

Колонка `tenant_id` в `chat_sessions` (по умолчанию `'default'`) — изоляция арендаторов.

---

## `007_audit_log.sql`

Таблица `audit_log` — действия админки (`kb_upload`, `kb_reindex`, login, …).  
Код: `internal/store/audit_store.go`, `internal/admin`.

---

## `008_reindex_jobs.sql`

Очередь асинхронного reindex (admin → worker).

---

## `012_ingest_jobs.sql`

Таблицы `ingest_jobs` / `ingest_tasks` — пайплайн parse → embed → index. См. [../INGESTION.md](../INGESTION.md).

---

## `009_pgvector.sql`

`CREATE EXTENSION vector` для `VECTOR_STORE=pgvector`.

---

## `010_saas_tenants.sql`

| Таблица | Назначение |
|---------|------------|
| `saas_tenants` | signup org, plan, Stripe customer |
| `tenant_quotas` | лимиты сообщений / storage / domains |
| `stripe_webhook_events` | идемпотентность webhook |

См. [../SAAS.md](../SAAS.md), [../BILLING.md](../BILLING.md).

---

## `011_admin_users_membership.sql`

| Таблица | Назначение |
|---------|------------|
| `admin_users` | admin-аккаунты (bcrypt, roles, tenant_id) |
| `user_tenant_memberships` | Telegram user ↔ tenant |

---

## Порядок файлов

```
001 … 003, 005 … 011  (нет 004)
```

---

## Как Go использует таблицы

| Таблица | Пример в коде |
|---------|----------------|
| `users` | `UpsertUser` в `internal/store/postgres_store.go` |
| `chat_sessions` | создание сессии, `domain_id`, `tenant_id` |
| `messages` | чат, поле `citations` |
| `message_feedback` | `internal/httpapi` + `internal/store` |
| `analytics_events` | `internal/store/analytics_store.go` |
| `audit_log` | `internal/store/audit_store.go` |
| `saas_tenants` | `internal/store/saas_store.go`, `internal/saas` |
| `admin_users` | `internal/admin/users_persist.go` |

---

## Краткий итог

| Файл | Что добавляет |
|------|----------------|
| **001** | users, chat_sessions, messages |
| **002** | `domain_id` |
| **003** | feedback, analytics_events |
| **005** | citations JSONB |
| **006** | `tenant_id` |
| **007** | audit_log |
| **008** | reindex jobs |
| **009** | pgvector extension |
| **010** | SaaS tenants / quotas / Stripe |
| **011** | admin_users, memberships |

Миграции — **версионированная схема БД**. Go применяет каждый файл один раз (`store.RunAllMigrations`) и пишет имя в `schema_migrations`.
