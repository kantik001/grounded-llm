# Roadmap — Grounded LLM

Strategy: **open platform for document-grounded assistants** — templates, API, on-prem deploy.  
Default language for docs and templates is **English**; Russian locale remains as a legacy locale pack.  
No country-specific logic in core — new languages ship as locale packs.

See also: [PLATFORM_VISION.md](../../PLATFORM_VISION.md) · [HIRING.md](../../HIRING.md)

---

## Vision

**Product:** platform to ship **cited, verified** internal document assistants from template packs in days.

**Positioning (one line):**

> *Open platform to deploy cited, verified document assistants in days — templates, API, on-prem.*

| We build | We do not build |
|----------|-----------------|
| Trust, data control, fast template rollout | “Another ChatGPT wrapper” |
| On-prem / private cloud first | Cloud-only lock-in |
| Measurable RAG quality (eval) | Unbounded agent workflows |

**Platform success metrics:**

| Metric | Target |
|--------|--------|
| New assistant from template | **&lt;3 days** (engineer-led) |
| Eval pass rate (retrieval) | Stable or improving per release |
| Fresh deploy | **&lt;1 day** |
| Community / adoption | Templates contributed, GitHub traction |

---

## Done (baseline)

### Platform core
- Document pipeline: `.txt`, `.pdf`, `.docx`
- Admin upload + reindex, citations UI, eval baseline
- Legacy API removed, `schema_migrations`

### Phase 1 — Trust
- Citations in chat, `rag_k`, admin index stats + delete, eval in CI

### Phase 2 — Integrators
- **SSE streaming** — `POST /message?stream=1`
- **API keys** — `X-API-Key`, `API_KEYS` / `API_KEYS_FILE`
- **API v1** — `/api/v1/*` + OpenAPI
- **Multi-tenant** — `X-Tenant-ID`, `data/{tenant}/{domain}/`, Chroma filter
- **Observability** — `X-Request-ID`, `/metrics`, structured logs
- **Admin feedback** — `GET /admin/feedback`
- **Domain scaffold** — `scripts/init_domain.sh` / `.ps1`

### i18n
- Docs `docs/en/` and `docs/ru/`
- Bundles: `config/locales/{ru,en}/`
- Middleware: `X-Locale`, `Accept-Language`, `?locale=`
- English API errors outside RU locale zone

**Summary:** strong **platform foundation**. Next: enterprise hardening (Phase B) and template catalog growth.

---

## Phase A — Platform baseline ✅

**Goal:** credible platform demo and reference template for integrators and hiring portfolio.

**Status:** complete in repository.

### Product

| Item | Status | Artifact |
|------|--------|----------|
| **English-first UI** | ✅ | `webapp/`, `DEFAULT_LOCALE=en` |
| **Security brief** | ✅ | [SECURITY_BRIEF.md](./SECURITY_BRIEF.md) |
| **HR reference template** | ✅ | [domain-packs/HR.md](./domain-packs/HR.md), `data/default/*_en.txt` |
| **Platform positioning** | ✅ | [PLATFORM_VISION.md](../../PLATFORM_VISION.md), [HIRING.md](../../HIRING.md) |
| **Locale extensibility** | ✅ | [LOCALE_GUIDE.md](./LOCALE_GUIDE.md) |

### Engineering

| Item | Status | Artifact |
|------|--------|----------|
| Webapp i18n | ✅ | `/branding`, locale bundles |
| Expand eval | ✅ | `eval/rag_default_en_baseline.jsonl` (18 cases) |
| Retrieval eval gate in CI | ✅ | job `eval-retrieval-gate` |
| Smoke E2E in CI | ✅ | job `smoke-api` |
| OpenAPI examples | ✅ | [API_EXAMPLES.md](./API_EXAMPLES.md) |

### Success criteria

- Fresh deploy: **&lt;1 day**
- Reference template runnable with eval pass
- README + vision clear in **&lt;60 seconds** for reviewers

---

## Phase B — Enterprise readiness (4–9 months)

**Goal:** pass security review and procurement at mid-market and enterprise.

### Product

| Item | Why |
|------|-----|
| **RBAC** | Roles: chat-only, KB editor, admin, API manager |
| **Audit log** | Upload, delete, reindex, admin login |
| **Per-tenant quotas** | Messages/day, storage, domains — billing foundation |
| **SSO (OIDC/SAML)** | Enterprise standard; Telegram stays optional |
| **Analytics dashboard** | Questions/day, verify pass rate, KB gaps, feedback |
| **Async reindex** | Job status instead of blocking admin |

### Engineering

| Item | Why |
|------|-----|
| Helm chart | Repeatable Kubernetes deploy |
| Backup/restore | Postgres + Chroma + `data/` |
| Readiness probes | postgres, python RAG, chroma separately |
| Retention policies | Configurable message/session retention |
| Retrieval improvements | Reranker or hybrid search; measured via eval |

### Adoption

- Template catalog growth (IT support, legal FAQ)
- Trust center: security, architecture, subprocessors
- Partner integrators (optional)

### Success criteria

- Security questionnaire answerable without per-client custom code
- Verify pass rate ≥75% on production eval
- External teams deploy from templates without forking core

---

## Phase C — Scalable platform (9–18 months)

**Goal:** repeatable rollouts without custom work per assistant.

### Product

| Item | Why |
|------|-----|
| **Hosted multi-tenant SaaS** | Signup → tenant → domain → upload (controlled beta) |
| **Billing** | Stripe/Paddle tied to quotas |
| **Plan tiers** | Starter / Business / Enterprise |
| **White-label light** | Logo, colors, app title via admin |
| **Embeddable widget** | Intranet embed, not only Telegram |
| **Managed vector DB** | Pinecone / Qdrant / pgvector |
| **Domain pack templates** | HR, IT support, legal FAQ |

### Engineering

| Item | Why |
|------|-----|
| Terraform modules | AWS / GCP / Azure |
| Multi-region deploy docs | Client chooses region |
| E2E eval with LLM | Quality gate on staging |
| SLA monitoring | Uptime and latency per tenant |

### Adoption

- Self-serve signup (controlled beta) for SMB
- Partner integrators ship on-prem for enterprise

### Success criteria

- New template pack in **&lt;3 days**
- Positive hosted tier margin (if enabled)
- Annual churn &lt;5% on supported deployments

---

## Phase D — Ecosystem & scale (18+ months)

**Goal:** partners and developers drive adoption, not only direct sales.

| Item | Why |
|------|-----|
| Webhooks / events | document indexed, verify failed |
| Ingest connectors | SharePoint, Google Drive, Confluence |
| Open core | MIT core vs commercial enterprise module |
| Partner certification | Integrator training program |
| Advanced analytics | Topics, KB gaps, article suggestions |
| Optional packs | Vision, support macros — separate SKUs |

### Success criteria

- ≥30% revenue through partners
- ≥5 production domains per tenant (Business tier)
- External contributions to domain packs

---

## Product principles (all phases)

1. **Grounded first** — RAG quality and verify beat feature count.
2. **Deploy anywhere** — on-prem, private cloud, hosted; same core.
3. **English default, locales pluggable** — no country logic in core.
4. **Domain pack = unit of rollout** — platform is enabler, template is the deliverable.
5. **Measure everything** — eval, metrics, feedback drive priorities.

---

## Explicitly out of scope (until demand)

- Country-specific compliance packages without a paying client
- New languages beyond EN (+ RU legacy) without a contract
- Consumer mobile app
- General chatbot without a knowledge base
- Per-client model fine-tuning (services exception only)

---

## At a glance

```text
NOW          Phase A              Phase B                 Phase C              Phase D
─────────────────────────────────────────────────────────────────────────────────────
Platform  →  templates + eval   → RBAC, audit, SSO,     → hosted tier +       → partners +
baseline      + positioning        connectors start        template catalog      marketplace
```

| Phase | Focus | Outcome |
|-------|-------|---------|
| **A** ✅ | Core + HR template + docs | Deployable platform |
| **B** | Enterprise hardening | Pass security review |
| **C** | Scale + optional SaaS | Repeatable rollouts |
| **D** | Ecosystem | Community-driven packs |

---

## Next priorities

**Phase A — done:**

1. ~~English-first webapp~~ ✅
2. ~~Security brief~~ ✅
3. ~~HR reference template + demo script~~ ✅
4. ~~Eval + smoke CI~~ ✅
5. ~~Platform vision + hiring portfolio docs~~ ✅

**Phase B — next:**

6. Minimal audit log
7. RBAC
8. ~~Retrieval eval gate in CI (Python + Chroma)~~ ✅
9. IT support template pack

---

## Relation to old «Phase 3»

The previous list (Helm, SaaS, vision pack, audit, dashboard) is **split across Phases B–D** and tied to platform maturity.

---

See also: [ARCHITECTURE.md](./ARCHITECTURE.md), [DEPLOY.md](./DEPLOY.md), [Russian ROADMAP](../ru/ROADMAP.md).
