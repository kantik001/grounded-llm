# Hosted SaaS (architecture prep)

**Status:** Phase 10 — signup API + UI scaffold; full hosted control plane remains Phase C.

Phase C goal: controlled multi-tenant signup with the same core as on-prem.

---

## Components

```text
Signup / billing (future)  →  tenant provisioner  →  existing stack
                                ├── Postgres (tenant row)
                                ├── data/{tenant}/
                                ├── config per tenant
                                └── Helm / Compose per tenant OR shared cluster
```

Reference implementation today: **single-tenant** Docker Compose / Helm.  
Multi-tenant data paths already exist (`X-Tenant-ID`, `data/{tenant}/`).

---

## Signup flow (implemented)

1. User opens `webapp/signup.html` or calls `POST /api/v1/signup`
2. Plan selected from `GET /api/v1/plans` ([plans.yaml](../../config/plans.yaml))
3. Tenant id allocated → written to `TENANTS_REGISTRY_FILE`
4. Quotas applied to `TENANT_QUOTAS_FILE`
5. Data directory `data/{tenant}/` created

Upgrade via Stripe Checkout (`POST /api/v1/billing/stripe/checkout`) + webhook.

Signup returns `admin_username` / `admin_password` once when `ADMIN_USERS_FILE` is configured.

---

## Environment

| Variable | Purpose |
|----------|---------|
| `SAAS_SIGNUP_ENABLED` | Gate public registration (`true` / `false`) |
| `TENANTS_REGISTRY_FILE` | Persisted tenant list (e.g. `config/tenants.json`) |
| `STRIPE_WEBHOOK_SECRET` | Verify Stripe webhook signatures |
| `PLANS_FILE` | Plan tiers (default `config/plans.yaml`) |

See [BILLING.md](./BILLING.md).

---

## Non-goals (until Phase C implementation)

- Multi-region control plane
- Per-tenant isolated clusters (enterprise only, manual)

---

## Related

- [ROADMAP.md](./ROADMAP.md) Phase C
- [config/QUOTAS.md](../../config/QUOTAS.md)
