# Hosted SaaS (architecture prep)

**Status:** Design document — Phase 9. Not implemented in reference deployment.

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

## Signup flow (target)

1. User registers (email + org)
2. Plan selected ([plans.yaml](../../config/plans.yaml))
3. Tenant id allocated → `ALLOWED_TENANTS`
4. Default domain + template pack install
5. Admin invites KB editors

---

## Environment (future)

| Variable | Purpose |
|----------|---------|
| `SAAS_SIGNUP_ENABLED` | Gate public registration |
| `STRIPE_SECRET_KEY` | Billing webhooks |
| `STRIPE_WEBHOOK_SECRET` | Verify events |

See [BILLING.md](./BILLING.md).

---

## Non-goals (until Phase C implementation)

- Multi-region control plane
- Per-tenant isolated clusters (enterprise only, manual)

---

## Related

- [ROADMAP.md](./ROADMAP.md) Phase C
- [config/QUOTAS.md](../../config/QUOTAS.md)
