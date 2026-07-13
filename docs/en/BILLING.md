# Billing & plan tiers (scaffold)

**Status:** Phase 10 — Stripe webhook + signup API implemented; **Checkout session creation is manual/Stripe Dashboard for now**.

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

## Integration (Phase 10–11)

1. **Signup** — `POST /api/v1/signup` creates tenant + applies plan quotas from `config/plans.yaml`
2. **Stripe Checkout** — `POST /api/v1/billing/stripe/checkout` for paid plans (`stripe_price_id` in plans.yaml)
3. **Webhook** — `POST /api/v1/billing/stripe/webhook` updates quotas on subscription events

Environment:

| Variable | Purpose |
|----------|---------|
| `SAAS_SIGNUP_ENABLED` | `true` to allow public signup |
| `TENANTS_REGISTRY_FILE` | e.g. `config/tenants.json` |
| `TENANT_QUOTAS_FILE` | quota enforcement file |
| `STRIPE_WEBHOOK_SECRET` | Stripe signing secret (`whsec_…`) |
| `STRIPE_SECRET_KEY` | Stripe API key for Checkout (`sk_…`) |
| `STRIPE_CHECKOUT_SUCCESS_URL` | Redirect after payment (default `http://localhost/signup.html?checkout=success`) |
| `STRIPE_CHECKOUT_CANCEL_URL` | Redirect on cancel |
| `PLANS_FILE` | defaults to `config/plans.yaml` |
| `ADMIN_USERS_FILE` | required for auto admin provision on signup |

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
