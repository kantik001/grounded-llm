# Hosted SaaS (optional beta layer)

**Status:** Phase 10–11 — signup API, Stripe Checkout + webhook, admin auto-provision. **Not required** for on-prem. Full multi-region control plane remains Phase C ([ROADMAP.md](./ROADMAP.md)).

---

## Components

```text
signup.html / POST /api/v1/signup
        ↓
tenant registry + quotas + data/{tenant}/
        ↓
optional Stripe Checkout → webhook → plan upgrade
        ↓
existing stack (Go + Python RAG + Postgres)
```

Reference deployment: **single-cluster** Docker Compose / Helm with multi-tenant paths (`X-Tenant-ID`, `data/{tenant}/`).

---

## Signup flow

1. User opens `webapp/signup.html` or calls `POST /api/v1/signup`
2. Plan from `GET /api/v1/plans` ([plans.yaml](../../config/plans.yaml))
3. Tenant id → `TENANTS_REGISTRY_FILE`
4. Quotas → `TENANT_QUOTAS_FILE`
5. Data dir `data/{tenant}/` created
6. Admin user `{tenant}-admin` when `ADMIN_USERS_FILE` is set (password returned once)
7. Paid plan → `checkout_url` when `STRIPE_SECRET_KEY` + `stripe_price_id` configured

Upgrade / renew via Stripe webhook ([BILLING.md](./BILLING.md)).

---

## Environment

| Variable | Purpose |
|----------|---------|
| `SAAS_SIGNUP_ENABLED` | Gate public registration (`true` / `false`) |
| `TENANTS_REGISTRY_FILE` | Persisted tenant list (e.g. `config/tenants.json`) |
| `TENANT_QUOTAS_FILE` | Enforced quotas per tenant |
| `ADMIN_USERS_FILE` | RBAC users file; signup appends tenant admin |
| `STRIPE_SECRET_KEY` | Checkout session API |
| `STRIPE_WEBHOOK_SECRET` | Verify webhook signatures |
| `PLANS_FILE` | Plan tiers (default `config/plans.yaml`) |

See [BILLING.md](./BILLING.md) for Checkout redirect URLs.

---

## Non-goals (until Phase C / hosted GA)

- Multi-region control plane
- Per-tenant isolated clusters (enterprise only, manual)
- Email delivery of admin credentials
- Stripe Customer Portal UI

---

## Related

- [ROADMAP.md](./ROADMAP.md)
- [LAUNCH.md](./LAUNCH.md)
- [config/QUOTAS.md](../../config/QUOTAS.md)
