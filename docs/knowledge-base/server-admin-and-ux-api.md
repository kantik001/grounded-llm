# Разбор: админка и UX API / Admin & UX API

**Файлы / Files:** `admin.go`, `domains.go`, `onboarding.go`, `branding.go`, `feedback.go`, `analytics_store.go`  
**Клиент / Client:** [webapp-overview.md](./webapp-overview.md)

---

## `admin.go` — knowledge base upload

### Авторизация

HTTP Basic: `ADMIN_USER` / `ADMIN_PASSWORD`. Пустой пароль → **503**.

### Маршруты `/admin` и `/api/admin`

| Метод | Handler | Действие |
|-------|---------|----------|
| GET | `handleAdminStatus` | `{ data_dir, domains }` |
| GET | `handleAdminListArticles` | файлы в `data/{domain_id}/` |
| POST | `handleAdminUpload` | сохранить документ |
| DELETE | `handleAdminDeleteArticle` | удалить документ (`?domain_id=&filename=`) |
| POST | `handleAdminReindex` | reindex через Python |

### `GET /admin/articles`

Ответ: `articles[]` с `filename`, `size_bytes`, `modified`, `chunks` (из Python `/admin/index-stats`).

### Upload

- `domain_id`
- Форматы: **`.txt`**, **`.pdf`**, **`.docx`**
- Regex: `^[a-zA-Z0-9._-]+\.(txt|pdf|docx)$`
- Max size: **10 МБ**
- Path: `{DATA_DIR}/{domain_id}/{filename}`

---

## `domains.go` — каталог доменов

`loadDomainCatalog()` ← `DOMAINS_CONFIG_PATH` / `config/domains.json`

### `GET /domains`, `/api/domains`

Публично, без Telegram auth.

---

## `onboarding.go`

`GET /onboarding?domain_id=default` → `{ questions: [...] }`

---

## `branding.go`

`GET /branding` → UI strings из `config/branding.json`

---

## `feedback.go`

`POST /feedback` — rating `1` / `-1` на ответ ассистента (Telegram auth).

---

## Reindex chain

Go `POST /admin/reindex` → Python `POST /admin/reindex` + header `X-Admin-Secret`.

См. [rag-vector_store.md](./rag-vector_store.md).
