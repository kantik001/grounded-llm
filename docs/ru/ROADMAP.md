# Дорожная карта — Grounded LLM

Стратегия: **международный B2B-продукт** — компании разворачивают ассистента по своим документам on-prem или в private cloud.  
Язык продаж и документации по умолчанию — **английский**; русская локаль остаётся для разработки и локального рынка.  
Привязки к конкретной стране нет — новые языки добавляются через locale packs, когда появится платящий спрос.

---

## Куда мы идём (vision)

**Продукт:** платформа для «заземлённых» ассистентов — ответы **только из базы знаний**, с цитатами и проверкой (verify).

**Как это продаётся за рубежом (одной фразой):**

> *Private AI assistant for internal documents — cited, verified, deployable in your infrastructure.*

| Мы продаём | Мы не продаём |
|------------|---------------|
| Доверие к ответам, контроль данных | «Ещё один ChatGPT» |
| Быстрое внедрение domain pack | Абстрактную «платформу без кейса» |
| On-prem / private cloud | Обязательный только облачный SaaS с первого дня |

**Как зарабатываем по фазам:**

| Фаза | Источник денег |
|------|----------------|
| A–B | Пилоты, внедрение «под ключ», годовая поддержка |
| C | Hosted multi-tenant + тарифы по подписке |
| D | Партнёры-интеграторы + enterprise-модули |

---

## Уже сделано (baseline)

### Ядро платформы
- Пайплайн документов: `.txt`, `.pdf`, `.docx`
- Админка: upload + reindex, citations в UI, eval baseline
- Удалён legacy API, учёт миграций в `schema_migrations`

### Фаза 1 — доверие к ответам
- Цитаты в чате, `rag_k`, статистика индекса и удаление статей в админке, eval в CI

### Фаза 2 — интеграторы
- **SSE streaming** — `POST /message?stream=1`
- **API keys** — `X-API-Key`, `API_KEYS` / `API_KEYS_FILE`
- **API v1** — `/api/v1/*` + OpenAPI
- **Мультитенантность** — `X-Tenant-ID`, `data/{tenant}/{domain}/`, фильтр Chroma
- **Наблюдаемость** — `X-Request-ID`, `/metrics`, структурированные логи
- **Admin feedback** — `GET /admin/feedback`
- **Scaffold домена** — `scripts/init_domain.sh` / `.ps1`

### Локализация (i18n)
- Документация `docs/ru/` и `docs/en/`
- Bundles: `config/locales/{ru,en}/`
- Middleware: `X-Locale`, `Accept-Language`, `?locale=`
- API-ошибки на английском (вне RU-зоны)

**Итог:** сильный **технический MVP** и зачаток **platform core**.  
Для международных продаж не хватает **продуктовой зрелости** (UI, security story, enterprise-фичи, упакованный vertical).

---

## Фаза A — продукт для международного рынка (0–4 месяца)

**Цель:** показать demo и запустить пилот без стыда — buyer или integrator понимает ценность за 30 минут.

### Продукт

| Задача | Зачем |
|--------|-------|
| **English-first UI** | Убрать русские fallback в webapp; по умолчанию `en` |
| **Security brief** | PDF на 2 стр.: куда идут данные, on-prem, LLM API, «не обучаем на ваших документах» |
| **Pilot playbook** | Шаблон SOW на 8 недель, KPI, сценарий demo |
| **HR domain pack (EN)** | Готовый vertical: промпты, onboarding, demo KB, one-pager для продаж |
| **Расширяемость locale** | Документ: как добавить новый язык без правок ядра |

### Инженерия

| Задача | Зачем |
|--------|-------|
| Довести i18n в webapp | Все строки UI из `/branding` и bundles, не из кода |
| Расширить eval | 15–25 EN-вопросов, gate в CI по retrieval |
| Smoke E2E в CI | `docker compose` + smoke — стабильность для integrators |
| Примеры OpenAPI | curl/Postman: session, message, stream, admin |

### Продажи и GTM (go-to-market)

- 2–3 разговора о пилоте (remote, на английском — можно с партнёром-переводчиком)
- 1 case study (можно анонимный)
- GitHub + `docs/en/` — главная точка входа для мира

### Критерии успеха фазы A

- Demo → пилот: конверсия ≥20% от заинтересованных
- На пилоте: ≥85% in-scope ответов **с цитатами**
- Развёртывание с нуля: **&lt;1 дня**

**Деньги на этой фазе:** пилот **$8k–25k** (разовый проект + отчёт).

---

## Фаза B — готовность к enterprise (4–9 месяцев)

**Цель:** пройти security review и закупку у компаний среднего и крупного размера.

### Продукт

| Задача | Зачем |
|--------|-------|
| **RBAC** | Роли: только чат / редактор KB / admin / управление API keys |
| **Audit log** | Кто загрузил/удалил документ, reindex, вход в админку |
| **Квоты per tenant** | Сообщения/день, объём KB, число доменов — основа для billing |
| **SSO (OIDC/SAML)** | Стандарт enterprise; Telegram остаётся опциональным каналом |
| **Analytics dashboard** | Вопросы/день, verify pass rate, пробелы в KB, feedback — UI для admin |
| **Async reindex** | Статус задачи, админ не «висит» 10+ минут |

### Инженерия

| Задача | Зачем |
|--------|-------|
| Helm chart | Повторяемый деплой в Kubernetes |
| Backup/restore | Postgres + Chroma + `data/` |
| Readiness probes | Отдельно postgres, python RAG, chroma |
| Retention policies | Настраиваемое хранение сообщений/сессий |
| Улучшение retrieval | Reranker или hybrid search; измерять через eval |

### GTM

- **Годовая лицензия** (не self-serve с карты)
- Partner program v1: 1–2 integrator, rev share
- Trust center: security, architecture, subprocessors (LLM API)

### Критерии успеха фазы B

- 1–2 оплаченные годовые лицензии
- Security questionnaire — без кастомного кода на каждого клиента
- Verify pass rate ≥75% на production eval
- Удовлетворённость админов пилота (NPS) ≥40

**Деньги:** годовая лицензия **$24k–80k** + support retainer.

---

## Фаза C — масштабируемая платформа и revenue (9–18 месяцев)

**Цель:** повторяемый доход без 100% ручной работы на каждого клиента.

### Продукт

| Задача | Зачем |
|--------|-------|
| **Hosted multi-tenant SaaS** | Регистрация → tenant → domain → upload (controlled beta) |
| **Billing** | Stripe/Paddle, привязка к квотам |
| **Тарифы** | Starter / Business / Enterprise |
| **White-label light** | Лого, цвета, название приложения через admin |
| **Embeddable widget** | Вставка в intranet (не только Telegram) |
| **Managed vector DB** | Pinecone / Qdrant / pgvector для масштаба |
| **Шаблоны domain packs** | HR, IT support, legal FAQ |

### Инженерия

| Задача | Зачем |
|--------|-------|
| Terraform modules | AWS / GCP / Azure |
| Multi-region deploy docs | Клиент выбирает регион |
| E2E eval с LLM | Quality gate на staging |
| SLA monitoring | Uptime и latency по tenant |

### GTM

- Self-serve для SMB
- Enterprise sales для on-prem и крупных контрактов

### Критерии успеха фазы C

- Положительный MRR с hosted tier
- Маржа на hosted (LLM + infra) в плюсе
- Churn годовых контрактов &lt;5%
- Новый domain pack за **&lt;3 дней**

---

## Фаза D — экосистема и масштаб (18+ месяцев)

**Цель:** продукт продают и внедряют партнёры, не только вы.

| Направление | Зачем |
|-------------|-------|
| Webhooks / events | document indexed, verify failed — интеграции |
| Коннекторы ingest | SharePoint, Google Drive, Confluence |
| Open core | MIT core vs commercial enterprise module |
| Partner certification | Обучение integrators |
| Advanced analytics | Темы вопросов, пробелы KB, подсказки новых статей |
| Optional packs | Vision, support macros — отдельные SKU |

### Критерии успеха фазы D

- ≥30% revenue через партнёров
- ≥5 production domains на tenant (Business tier)
- Внешние contribution в domain packs

---

## Принципы продукта (на все фазы)

1. **Grounded first** — качество RAG и verify важнее количества фич.
2. **Deploy anywhere** — on-prem, private cloud, hosted; одно ядро.
3. **English default, locales pluggable** — без логики «под страну» в core.
4. **Domain pack = единица продажи** — платформа enabler, vertical — оффер.
5. **Measure everything** — eval, metrics, feedback задают приоритеты.

---

## Сознательно не делаем (пока нет спроса)

- Compliance «под страну» без платящего клиента
- Новые языки кроме EN (+ RU для legacy) без контракта
- Consumer mobile app
- Chatbot без привязки к KB
- Fine-tuning моделей per client (только как услуга)

---

## Сводная таблица

```text
СЕЙЧАС     Фаза A                 Фаза B                  Фаза C               Фаза D
──────────────────────────────────────────────────────────────────────────────────────
MVP+i18n → EN product + HR pack → RBAC, audit, SSO,     → SaaS + billing +    → Partners +
           пилоты + eval          dashboard               white-label           connectors
```

| Фаза | Срок | Кто покупает | Модель денег |
|------|------|--------------|--------------|
| **A** | 0–4 мес | HR / IT (пилот) | Пилот $8k–25k |
| **B** | 4–9 мес | CISO + HR + закупки | Лицензия $24k–80k/год |
| **C** | 9–18 мес | SMB + enterprise | MRR + enterprise |
| **D** | 18+ мес | Партнёры | License + marketplace |

---

## Ближайшие 90 дней (конкретный backlog)

**Обязательно в продукте:**

1. English-first webapp (без RU fallback)
2. Security & architecture brief (EN PDF)
3. HR domain pack: one-pager + demo script
4. Расширить eval + smoke в CI
5. Audit log (минимальный) — даже до полного RBAC
6. Смержить i18n-ветку, тег `v0.9` «international beta»

**Обязательно в бизнесе:**

1. 10 исходящих контактов (LinkedIn, EN)
2. 1 платный или оплачиваемый пилот
3. 1 разговор с integrator

---

## Связь со старым «Фаза 3»

Прежний список (Helm, SaaS, vision pack, audit, dashboard) **разложен по фазам B–D** и привязан к деньгам и покупателям.  
**Фаза A** — новый обязательный шаг: без него международный рынок не откроется, даже при хорошем коде.

---

См. также: [ARCHITECTURE.md](./ARCHITECTURE.md), [DEPLOY.md](./DEPLOY.md), [английская версия ROADMAP](../en/ROADMAP.md).
