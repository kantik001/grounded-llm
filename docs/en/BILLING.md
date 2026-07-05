# Billing & plan tiers (scaffold)

**Status:** Phase 9 scaffold — `config/plans.yaml` defines tiers; **no Stripe integration yet**.

Maps to existing per-tenant quotas ([config/QUOTAS.md](../../config/QUOTAS.md)).

---

## Plans

| Plan | Price | Messages/day | Storage | Domains |
|------|-------|--------------|---------|---------|
| **Starter** | Free | 200 | 512 MB | 1 |
| **Business** | $299/mo | 5,000 | 10 GB | 10 |
| **Enterprise** | Contact sales | Custom | Custom | Custom |

Source of truth: [config/plans.yaml](../../config/plans.yaml).

---

## Integration path (Phase 10+)

1. **Stripe Checkout** — create subscription on signup
2. **Webhook** — `customer.subscription.updated` → update `config/tenant_quotas.json`
3. **Enforcement** — existing Go quota middleware ([QUOTAS.md](../../config/QUOTAS.md))

Suggested webhook events:

- `checkout.session.completed`
- `customer.subscription.deleted`
- `invoice.payment_failed`

---

## On-prem / OSS

Plans file is **documentation only** for self-hosted deployments.  
Operators set quotas directly in `tenant_quotas.json` without billing.

---

## Related

- [SAAS.md](./SAAS.md)
- [PHASE_9.md](./PHASE_9.md)
