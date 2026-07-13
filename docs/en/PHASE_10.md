# Phase 10 — SaaS billing & self-serve signup

**Goal:** Stripe webhook → tenant quotas, public signup API + UI, tenant registry.

**Branch:** `feature/phase-10-saas-billing`  
**Horizon:** 3 (hosted beta prep)  
**Prerequisite:** Phase 9 merged to `main` ✅

---

## Pillars addressed

| Pillar | Phase 10 deliverables |
|--------|------------------------|
| **3 Platform** | Self-serve signup API + `webapp/signup.html` |
| **Path B** | Stripe webhook → `tenant_quotas.json` |
| **Adoption** | Tenant registry (`config/tenants.json`) |

---

## Deliverables

| # | Item | Artifact |
|---|------|----------|
| 1 | Signup API | `POST /api/v1/signup` |
| 2 | Plans API | `GET /api/v1/plans` |
| 3 | Stripe webhook | `POST /api/v1/billing/stripe/webhook` |
| 4 | Tenant registry | `config/tenants.json.example`, `server/tenant_registry.go` |
| 5 | Signup UI | `webapp/signup.html` |
| 6 | Docs | [BILLING.md](./BILLING.md), [SAAS.md](./SAAS.md) |

---

## Acceptance criteria

### Enable signup (dev)
```bash
export SAAS_SIGNUP_ENABLED=true
export TENANTS_REGISTRY_FILE=config/tenants.json
export TENANT_QUOTAS_FILE=config/tenant_quotas.json
cp config/tenants.json.example config/tenants.json
```

### Signup
```bash
curl -X POST http://localhost:8080/api/v1/signup \
  -H 'Content-Type: application/json' \
  -d '{"org_name":"Acme","email":"admin@acme.com","plan":"starter"}'
```

### Stripe webhook (local with Stripe CLI)
```bash
stripe listen --forward-to localhost:8080/api/v1/billing/stripe/webhook
```

### UI
Open `http://localhost/signup.html` (nginx webapp) when signup is enabled.

---

## Out of scope (Phase 11+)

- Stripe Checkout session creation endpoint
- Self-serve admin user provisioning
- Making repository public (operator decision — [LAUNCH.md](./LAUNCH.md))

---

## Related

- [PHASE_9.md](./PHASE_9.md)
- [BILLING.md](./BILLING.md)
- [SAAS.md](./SAAS.md)
