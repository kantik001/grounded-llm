# Авторизация и лимиты (`server/`)

**Файлы:** `auth_telegram.go`, `api_keys.go`, `auth_combined.go`, `middleware.go`, `ratelimit.go`, `locale.go`  
**Связь:** [webapp-overview.md](./webapp-overview.md), [server-overview.md](./server-overview.md)

---

## Типы доступа

| Тип | Кто | Как |
|-----|-----|-----|
| **Пользователь** | Telegram Web App | `X-Telegram-Init-Data` → HMAC-проверка |
| **Интегратор** | HTTP-клиенты, боты | `X-API-Key` (`API_KEYS` / `API_KEYS_FILE`) |
| **Админ** | Браузер `/admin.html` | HTTP Basic `ADMIN_USER` / `ADMIN_PASSWORD` |

Эта статья — про **Telegram + API keys + CORS + locale + rate limit**.

---

## `auth_telegram.go` — проверка Telegram

### Что такое `initData`

Строка query-параметров от Telegram Web App (`user=...&auth_date=...&hash=...`).  
Подписана секретом бота — **подделать без токена бота нельзя**.

Документация: [Telegram Web Apps — validating data](https://core.telegram.org/bots/webapps#validating-data-received-via-the-mini-app).

### Алгоритм `validateTelegramInitData`

1. Разбор query, извлечение `hash`.
2. Сборка `data_check_string` (остальные поля, сортировка).
3. HMAC-SHA256 с ключом из `botToken` + `"WebAppData"`.
4. Сравнение с `hash` (constant-time).
5. Проверка `auth_date` — не старше `TELEGRAM_INIT_DATA_MAX_AGE_SEC` (по умолчанию 86400).
6. Парсинг JSON поля `user` → `TelegramUser` (id, имя, username, **`language_code`** для локали).

Тесты: `auth_telegram_test.go`.

---

## `api_keys.go` — ключи интеграторов

- Заголовок `X-API-Key` сверяется с `API_KEYS` (через запятую) или файлом `API_KEYS_FILE`.
- Используется в `combinedAuthMiddleware` вместе с Telegram auth.
- OpenAPI: `/api/v1/openapi.json`.

---

## `middleware.go` — CORS и locale

### CORS

- `CORS_ALLOWED_ORIGINS` (через запятую).
- Методы GET, POST, OPTIONS.
- Разрешённые заголовки: `Content-Type`, `X-Telegram-Init-Data`, `Authorization`, `X-API-Key`, `X-Tenant-ID`, `X-Locale`, `Accept-Language`.

### Locale

`localeMiddleware` определяет язык запроса:

1. `X-Locale`
2. query `?locale=`
3. `Accept-Language`
4. `language_code` из Telegram user
5. `DEFAULT_LOCALE` (по умолчанию `ru`)

---

## `combinedAuthMiddleware`

**Режим разработки** (`TELEGRAM_AUTH_DISABLED=true`):

- Пропускает проверку initData.
- В контекст кладёт `telegram_user_id` = 1 (или `X-Dev-User-Id`).

**Продакшен:**

1. Валидный `X-API-Key` → доступ (синтетический user id для лимита).
2. Иначе: `X-Telegram-Init-Data` или `Authorization: tma <initData>`.
3. Пусто → **401** («откройте из бота»).
4. Неверная подпись → **401**.

Handlers: `ctxActorUser(c)` / `ctxTelegramUser(c)`.

---

## Видимость маршрутов

**Публичные (без chat auth):** `/health`, `/ready`, `/domains`, `/onboarding`, `/branding`, `/metrics`

**SaaS (опц., `SAAS_SIGNUP_ENABLED=true`):** `GET /api/v1/plans`, `POST /api/v1/signup`, `POST /api/v1/billing/stripe/checkout`, webhook — [SAAS.md](../SAAS.md)

**Админка:** `/api/admin/*` — Basic, OIDC или `ADMIN_USERS_FILE`

---

## Защищённые маршруты (chat API)

С **`auth` + rate limit**:

- `/session`, `/history`, `/message`, `/feedback`
- Дубли: `/api/...`, `/api/v1/...`

---

## `ratelimit.go` — лимит запросов

- `RATE_LIMIT_REQUESTS_PER_MINUTE` (по умолчанию 30).
- Окно 1 минута, in-memory по user id.
- `0` → лимит выключен.
- Превышение → **429** «Слишком много запросов…».

При нескольких репликах Go счётчики не общие (в перспективе — Redis).

---

## Типичные ошибки

| Симптом | Причина |
|---------|---------|
| 401 на session | нет initData/API key, auth не отключён |
| 401 «неверная подпись» | неверный `TELEGRAM_BOT_TOKEN` |
| 401 «устарел» | старый initData — переоткройте Web App |
| 401 API key | неверный или отсутствует `X-API-Key` |
| 429 | больше N запросов в минуту с одного user id |

---

## Краткий итог

Telegram и API keys — два способа доступа к чату. CORS и locale — для Web App. Rate limit защищает LLM от спама.
