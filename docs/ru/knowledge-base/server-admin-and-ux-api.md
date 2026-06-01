# Админка и UX API

**Файлы:** `admin.go`, `admin_feedback.go`, `domains.go`, `onboarding.go`, `branding.go`, `feedback.go`, `locale.go`  
**Клиент:** [webapp-overview.md](./webapp-overview.md)

---

## `admin.go` — загрузка базы знаний

### Авторизация

HTTP Basic: `ADMIN_USER` / `ADMIN_PASSWORD`. Пустой пароль → **503**.

### Маршруты `/admin` и `/api/admin`

| Метод | Обработчик | Действие |
|-------|---------|----------|
| GET | `handleAdminStatus` | `{ data_dir, domains }` |
| GET | `handleAdminListArticles` | файлы в `data/{tenant}/{domain}/` |
| POST | `handleAdminUpload` | сохранить документ |
| DELETE | `handleAdminDeleteArticle` | удалить документ |
| POST | `handleAdminReindex` | reindex через Python |
| GET | `handleAdminFeedbackSummary` | сводка 👍/👎 |

### Upload

- `domain_id`, опционально `tenant_id`
- Форматы: **`.txt`**, **`.pdf`**, **`.docx`**
- Имя файла: латиница, цифры, `_`, `-`; до **10 МБ**
- Путь: `{DATA_DIR}/{tenant_id}/{domain_id}/{filename}`

---

## `domains.go` — каталог доменов

`loadDomainCatalog()` ← `config/domains.json`

### `GET /domains`

Публично, без Telegram auth. Имя домена — по локали запроса (`names.ru` / `names.en`).  
В ответе поле `locale`.

---

## `onboarding.go`

`GET /onboarding?domain_id=default&locale=ru` → `{ questions: [...], locale }`

Данные из `config/locales/{locale}/onboarding.json`.

---

## `branding.go`

`GET /branding?locale=ru` → строки UI из `config/locales/{locale}/branding.json`

---

## `feedback.go`

`POST /feedback` — оценка `1` / `-1` на сообщение ассистента (Telegram или API key).

---

## Цепочка reindex

Go `POST /admin/reindex` → Python `POST /admin/reindex` + `X-Admin-Secret`.

→ [rag-vector_store.md](./rag-vector_store.md)

---

## API для интеграторов (фаза 2)

- `/api/v1/*` — версионированные маршруты + OpenAPI
- `GET /metrics` — метрики Prometheus

→ [server-auth-and-limits.md](./server-auth-and-limits.md)
