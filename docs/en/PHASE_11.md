# Phase 11 — Checkout & admin provisioning

**Goal:** Stripe Checkout session API, self-serve admin user on signup, paid-plan payment flow.

**Branch:** `feature/phase-11-checkout-admin`  
**Horizon:** 3 (hosted beta)  
**Prerequisite:** Phase 10 merged to `main` ✅

---

## Pillars addressed

| Pillar | Phase 11 deliverables |
|--------|------------------------|
| **Path B** | `POST /api/v1/billing/stripe/checkout` |
| **3 Platform** | Admin user auto-provision on signup |
| **Adoption** | Signup UI shows admin creds + Stripe pay link |

---

## Deliverables

| # | Item | Artifact |
|---|------|----------|
| 1 | Checkout API | `server/stripe_checkout.go` |
| 2 | Admin provision | `server/admin_users_persist.go` |
| 3 | Plan Stripe price | `stripe_price_id` in `config/plans.yaml` |
| 4 | Signup UX | `webapp/signup.html` admin + checkout blocks |

---

## Acceptance criteria

### Checkout (existing tenant)
```bash
export SAAS_SIGNUP_ENABLED=true
export STRIPE_SECRET_KEY=sk_test_...
export PLANS_FILE=config/plans.yaml  # business.stripe_price_id=price_xxx

curl -X POST http://localhost:8080/api/v1/billing/stripe/checkout \
  -H 'Content-Type: application/json' \
  -d '{"tenant_id":"acme-demo","plan":"business","email":"a@b.com"}'
```

### Signup with admin user
```bash
export ADMIN_USERS_FILE=config/admin_users.json
curl -X POST http://localhost:8080/api/v1/signup \
  -H 'Content-Type: application/json' \
  -d '{"org_name":"Acme","email":"admin@acme.com","plan":"starter"}'
# → admin_username, admin_password (once)
```

### Paid plan flow
- Signup with `business` + Stripe configured → `checkout_url`, starter quotas until webhook upgrades plan

---

## Out of scope (Phase 12+)

- Email delivery of admin credentials
- Making repository public ([LAUNCH.md](./LAUNCH.md) — operator decision)
- Stripe Customer Portal / invoice UI

---

## Related

- [PHASE_10.md](./PHASE_10.md)
- [BILLING.md](./BILLING.md)
- [LAUNCH.md](./LAUNCH.md)
