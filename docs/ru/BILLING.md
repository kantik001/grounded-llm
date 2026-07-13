# Billing и тарифы

Краткая русская версия. **Канон:** [BILLING.md (EN)](../en/BILLING.md).

**Статус:** фазы 10–11 реализованы в коде; для on-prem billing **не нужен**.

---

## Тарифы (`config/plans.yaml`)

| План | Цена | Сообщений/день | Хранилище | Домены |
|------|------|----------------|-----------|--------|
| Starter | Бесплатно | 200 | 512 MB | 1 |
| Business | $299/мес | 5 000 | 10 GB | 10 |
| Enterprise | По запросу | Custom | Custom | Custom |

---

## Поток

1. Signup → tenant + квоты  
2. Checkout (`POST /api/v1/billing/stripe/checkout`) для Business  
3. Webhook обновляет квоты после оплаты  

До webhook на платном плане действуют **starter-квоты**.

---

## On-prem

Задайте квоты напрямую в `tenant_quotas.json` — см. [config/QUOTAS.md](../../config/QUOTAS.md).
