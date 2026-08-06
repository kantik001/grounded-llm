**Канон (EN):** [EMBED.md](../en/EMBED.md)

# Встраиваемый чат-виджет

Лёгкий **iframe-friendly** UI для интранет-порталов — те же `/api/session` и `/api/message`, что и основной webapp.

Зачем: виджет на портале без отдельного SPA; оркестрация остаётся на Go `:8080`.

---

## Quick start

1. Разверните Grounded LLM (Compose или K8s).
2. Разрешите origin портала в `CORS_ALLOWED_ORIGINS` на Go-сервере.
3. Вставьте:

```html
<iframe
  src="https://your-host/embed.html?api=/api/&tenant=default&locale=ru"
  width="420"
  height="560"
  style="border:0;border-radius:12px"
  title="Grounded assistant"
></iframe>
```

Локально: `http://localhost/embed.html?api=/api/`

---

## Query parameters

| Параметр | Default | Описание |
|----------|---------|----------|
| `api` | `/api/` | Base path API (прокси на Go) |
| `tenant` | `default` | `X-Tenant-ID` |
| `locale` | `en` | `X-Locale` |

---

## Security

- `embed.html` использует ослабленный `frame-ancestors *` CSP — **сузьте для prod** (только ваши интранет-домены).
- Включите Telegram auth или API keys (`TELEGRAM_AUTH_DISABLED`, `API_KEYS`) по политике деплоя.
- Не отдавайте admin-роуты с того же origin без edge ACL.

---

## Файлы

| Файл | Роль |
|------|------|
| `webapp/embed.html` | Shell |
| `webapp/embed.css` | Компактный layout |
| `webapp/embed.js` | Session + message |

Nginx: `webapp/nginx.conf`, location `=/embed.html`.

---

## См. также

- [DEPLOY.md](./DEPLOY.md)
- [NETWORK_SECURITY.md](./NETWORK_SECURITY.md)
