# Hosted SaaS (опциональный слой)

Краткая русская версия. **Канон:** [SAAS.md (EN)](../en/SAAS.md).

**Статус:** фазы 10–11 — signup API, Stripe Checkout + webhook, автосоздание admin. **Не обязательно** для on-prem.

---

## Компоненты

```
signup.html / POST /api/v1/signup
    → реестр tenant + квоты + data/{tenant}/
    → опционально Stripe Checkout → webhook
    → существующий стек (Go + Python RAG + Postgres)
```

---

## Поток регистрации

1. `webapp/signup.html` или `POST /api/v1/signup`
2. План из `GET /api/v1/plans` (`config/plans.yaml`)
3. Tenant → `TENANTS_REGISTRY_FILE`
4. Квоты → `TENANT_QUOTAS_FILE`
5. Каталог `data/{tenant}/`
6. Admin `{tenant}-admin` при `ADMIN_USERS_FILE` (пароль один раз в ответе)
7. Платный план → `checkout_url` при настроенном Stripe

---

## Переменные

| Переменная | Назначение |
|------------|------------|
| `SAAS_SIGNUP_ENABLED` | `true` — публичная регистрация |
| `TENANTS_REGISTRY_FILE` | `config/tenants.json` |
| `TENANT_QUOTAS_FILE` | файл квот |
| `ADMIN_USERS_FILE` | RBAC-пользователи |
| `STRIPE_SECRET_KEY` / `STRIPE_WEBHOOK_SECRET` | Checkout + webhook |
| `PLANS_FILE` | тарифы (по умолчанию `config/plans.yaml`) |

Подробнее: [BILLING.md](./BILLING.md) · [config/QUOTAS.md](../../config/QUOTAS.md)

---

## Не в scope (до hosted GA)

- Multi-region control plane
- Email с паролем admin
- Stripe Customer Portal
