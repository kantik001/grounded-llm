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
| POST | `handleAdminReindex` | reindex через Python |

### Upload

- `domain_id` (legacy form field: `crop_id`)
- Форматы: **`.txt`**, **`.pdf`**, **`.docx`**
- Regex: `^[a-zA-Z0-9._-]+\.(txt|pdf|docx)$`
- Max size: **10 МБ**
- Path: `{DATA_DIR}/{domain_id}/{filename}`

---

## `domains.go` — каталог доменов

`loadDomainCatalog()` ← `DOMAINS_CONFIG_PATH` / `config/domains.json`

### `GET /domains`, `/api/domains`

Публично, без Telegram auth.

Legacy: `GET /crops` — тот же каталог в старом формате JSON.

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
