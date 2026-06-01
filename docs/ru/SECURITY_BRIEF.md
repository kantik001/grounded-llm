# Security Brief — Grounded LLM

**Для кого:** IT-безопасность, закупки, compliance  
**Версия:** Фаза A (международный продукт)

---

## Кратко

Grounded LLM — ассистент по **вашим документам**, разворачивается **on-prem или в private cloud**. Ответы строятся только из загруженной базы знаний. Платформа **не обучает** foundation-модели на ваших данных.

| Свойство | Детали |
|----------|--------|
| Деплой | Docker Compose или Kubernetes (инфраструктура клиента) |
| Данные | Документы, чат и индекс остаются **у клиента** |
| LLM | Опционально внешний OpenAI-compatible API; в промпт попадают только фрагменты из RAG |
| Авторизация | Telegram initData, API keys, Basic Auth в админке |

---

## Поток данных

```text
Пользователь (Web / Telegram / API)
        │
        ▼
   Go server ──► PostgreSQL (сессии, сообщения, feedback)
        │
        ├──► Python RAG ──► Chroma (векторный индекс)
        │         ▲
        │         └── читает data/{tenant}/{domain}/
        │
        └──► LLM API (HTTPS) — вопрос + найденный контекст
```

**Что может уходить наружу** (если LLM в облаке):

- Текст вопроса
- **Фрагменты** документов (chunks) как контекст
- Системные промпты из ваших locale bundles

**Что остаётся внутри:**

- Полный корпус документов
- Chroma и Postgres
- История чата

Fine-tuning на документах клиента **не выполняется**.

---

## Компоненты и хранение

| Компонент | Что хранит |
|-----------|------------|
| PostgreSQL | Пользователи, сессии, сообщения, feedback |
| Chroma | Embeddings и метаданные chunks |
| `data/` | Файлы KB (`.txt`, `.pdf`, `.docx`) |
| `UPLOAD_DIR` | Загруженные изображения (опционально) |

Изоляция арендаторов: `tenant_id` в сессиях и фильтре Chroma.

---

## Доступ и аутентификация

| Интерфейс | Механизм |
|-----------|----------|
| Чат / Web App | Подпись Telegram `initData` |
| Интеграторы | `X-API-Key` |
| Админка KB | Basic Auth |
| Reindex (Python) | `X-Admin-Secret` |

**Для production:**

- Сильный `ADMIN_PASSWORD` и `ADMIN_SECRET`
- `TELEGRAM_AUTH_DISABLED=false`
- Ограничить `CORS_ALLOWED_ORIGINS`
- Admin и `/metrics` — за VPN или ACL proxy
- Ротация API keys

---

## Субпроцессоры (опционально)

При внешнем LLM (OpenRouter, OpenAI, Azure OpenAI) провайдер получает промпт (контекст + вопрос). Изучите DPA провайдера. Для air-gap — локальный OpenAI-compatible endpoint.

Embeddings (`multilingual-e5-small`) работают **внутри Python-контейнера**, без внешнего embedding API.

---

## Логи

- `X-Request-ID` в запросах
- `[RAG]` — domain, verify, без полного текста LLM
- `/metrics` — защитить в production

Полный audit log UI — **Фаза B**.

---

## Чеклист перед prod

- [ ] TLS на reverse proxy
- [ ] Секреты не в git
- [ ] Backup Postgres, Chroma, `data/`
- [ ] Rate limits
- [ ] Ротация LLM API key
- [ ] Dev-флаги выключены

---

См. также: [ARCHITECTURE.md](./ARCHITECTURE.md), [DEPLOY.md](./DEPLOY.md), [PILOT_PLAYBOOK.md](./PILOT_PLAYBOOK.md).
