# Дорожная карта — Grounded LLM

> **Фазы доставки 1–11 (spec, connectors, SaaS):** сводка ниже; детали — [docs/en/ROADMAP.md](../en/ROADMAP.md).  
> Стратегические фазы **A–D** — долгосрочный план (продажи, enterprise, экосистема).

Стратегия: **open platform** — cited, verified document assistants on-prem / private cloud.  
Язык продукта и спецификации по умолчанию — **английский**; `docs/ru/` — для RU-локали и пилотов.  
Новые языки — через locale packs, без логики «под страну» в core.

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

## Фаза A — продукт для международного рынка (0–4 месяца) ✅

**Цель:** показать demo и запустить пилот без стыда — buyer или integrator понимает ценность за 30 минут.

**Статус:** deliverables Фазы A реализованы в репозитории (смержены в `main`).

### Продукт

| Задача | Статус | Артефакт |
|--------|--------|----------|
| **English-first UI** | ✅ | `webapp/`, `DEFAULT_LOCALE=en` |
| **Security brief** | ✅ | [SECURITY_BRIEF.md](./SECURITY_BRIEF.md) |
| **Case study template** | ✅ | [CASE_STUDY_HR_PILOT.md](../en/CASE_STUDY_HR_PILOT.md) |
| **HR domain pack (EN)** | ✅ | [domain-packs/HR.md](./domain-packs/HR.md), `data/default/*_en.txt` |
| **Расширяемость locale** | ✅ | [LOCALE_GUIDE.md](./LOCALE_GUIDE.md) |

### Инженерия

| Задача | Статус | Артефакт |
|--------|--------|----------|
| i18n в webapp | ✅ | `/branding`, locale bundles |
| Расширить eval | ✅ | `eval/rag_default_en_baseline.jsonl` (18 вопросов) |
| Smoke E2E в CI | ✅ | job `smoke-api` в `.github/workflows/ci.yml` |
| Примеры OpenAPI | ✅ | [API_EXAMPLES.md](./API_EXAMPLES.md) |

### Продажи и GTM (go-to-market)

- 2–3 разговора о пилоте (remote, на английском — можно с партнёром-переводчиком)
- 1 case study (можно анонимный)
- GitHub + `docs/en/` — главная точка входа для мира

### Критерии успеха фазы A

- Demo → пилот: конверсия ≥20% от заинтересованных
- На пилоте: ≥85% in-scope ответов **с цитатами**
- Развёртывание с нуля: **&lt;1 дня**

---

## Фаза B — enterprise readiness (4–9 месяцев)

**Цель:** security review и закупка у mid-market / enterprise.

**Большая часть инженерных пунктов уже в `main`:** RBAC, audit log, квоты, OIDC SSO, analytics dashboard, async reindex, Helm, retention, hybrid/cross-encoder rerank. См. [ROADMAP (EN)](../en/ROADMAP.md) Phase B checklist.

Остаётся в основном **GTM и SAML**, trust center под конкретного клиента.

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

| Фаза | Срок | Кто покупает | Модель |
|------|------|--------------|--------|
| **A** | 0–4 мес | HR / IT (пилот) | Проектный пилот + KPI-отчёт |
| **B** | 4–9 мес | CISO + HR + закупки | Годовая лицензия on-prem |
| **C** | 9–18 мес | SMB + enterprise | MRR + enterprise |
| **D** | 18+ мес | Партнёры | License + marketplace |

---

## Фазы доставки 1–11 ✅ (сводка)

Нумерованные фазы в коде **завершены** (смержены в `main`).

| Фаза | Суть |
|------|------|
| **1–3** | Trust, API v1, enterprise (Helm, readiness, retention) |
| **4** | Spec, conformance, tenant purge |
| **5** | Spec v1, conformance CLI, site, RFC |
| **6** | Legal FAQ pack, vector adapters, hybrid, AWS Terraform |
| **7** | Pack registry, cross-encoder, connectors, GCP, governance |
| **8** | SharePoint Graph, Azure TF, embed widget, export connectors |
| **9** | Live Drive/Confluence, launch playbook, plans scaffold |
| **10** | Signup API, Stripe webhook, tenant registry |
| **11** | Stripe Checkout, admin auto-provision on signup |

Детали: [PHASE_4.md](../en/PHASE_4.md) … [PHASE_11.md](../en/PHASE_11.md) (EN).

---

## Что дальше (после фазы 11)

Не обязательная «фаза 12 в коде» — выбор оператора:

| Трек | Фокус | Документ |
|------|-------|----------|
| **Launch** | Public repo, `v0.3.0`, Pages, dev.to / HN | [LAUNCH.md](./LAUNCH.md) |
| **Hosted beta** | Email creds, Stripe Portal, staging | [SAAS.md](./SAAS.md) |
| **Enterprise pilot** | SAML, trust center, первый клиент | [PARTNER_CERTIFICATION.md](../en/PARTNER_CERTIFICATION.md), [TRUST_CENTER.md](../en/TRUST_CENTER.md) |

---

## Связь со старым «Фаза 3»

Прежний список (Helm, SaaS, vision pack, audit, dashboard) **разложен по фазам B–D и 1–11**.  
**Фаза A** — обязательный фундамент для международного позиционирования.

---

См. также: [ARCHITECTURE.md](./ARCHITECTURE.md), [DEPLOY.md](./DEPLOY.md), [README.md](./README.md), [английская ROADMAP](../en/ROADMAP.md).
