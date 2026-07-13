# Billing & plan tiers

**Status:** Phase 10–11 — signup, Stripe Checkout, and webhook → tenant quotas. **Optional** for on-prem; enable only for hosted beta.

Maps to per-tenant quotas ([config/QUOTAS.md](../../config/QUOTAS.md)).

---

## Plans

| Plan | Price | Messages/day | Storage | Domains |
|------|-------|--------------|---------|---------|
| **Starter** | Free | 200 | 512 MB | 1 |
| **Business** | $299/mo | 5,000 | 10 GB | 10 |
| **Enterprise** | Contact sales | Custom | Custom | Custom |

Source of truth: [config/plans.yaml](../../config/plans.yaml) (`stripe_price_id` for paid Checkout).

---

## Integration flow

1. **Signup** — `POST /api/v1/signup` creates tenant, provisions admin (if `ADMIN_USERS_FILE` set), applies quotas
2. **Paid plan** — `POST /api/v1/billing/stripe/checkout` or `checkout_url` returned from signup when Stripe is configured
3. **Webhook** — `POST /api/v1/billing/stripe/webhook` upgrades plan quotas on `checkout.session.completed`

Paid signup applies **starter quotas** until webhook confirms payment.

---

## Environment

| Variable | Purpose |
|----------|---------|
| `SAAS_SIGNUP_ENABLED` | `true` to allow public signup |
| `TENANTS_REGISTRY_FILE` | e.g. `config/tenants.json` |
| `TENANT_QUOTAS_FILE` | quota enforcement file |
| `ADMIN_USERS_FILE` | auto-create `{tenant}-admin` on signup |
| `STRIPE_SECRET_KEY` | Stripe API key for Checkout (`sk_…`) |
| `STRIPE_WEBHOOK_SECRET` | Stripe signing secret (`whsec_…`) |
| `STRIPE_CHECKOUT_SUCCESS_URL` | Redirect after payment |
| `STRIPE_CHECKOUT_CANCEL_URL` | Redirect on cancel |
| `PLANS_FILE` | defaults to `config/plans.yaml` |

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
- [PHASE_10.md](./PHASE_10.md)
- [PHASE_11.md](./PHASE_11.md)
