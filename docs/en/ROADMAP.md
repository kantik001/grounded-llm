# Roadmap — Grounded LLM

Strategy: **international B2B product** — organizations deploy a document-grounded assistant on-prem or in private cloud.  
Default language for sales and docs is **English**; Russian locale remains for local development and market.  
No country-specific features in core — new languages ship as locale packs when there is paying demand.

---

## Vision

**Product:** platform core for grounded assistants — answers **only from the knowledge base**, with citations and verification.

**International positioning (one line):**

> *Private AI assistant for internal documents — cited, verified, deployable in your infrastructure.*

| We sell | We do not sell |
|---------|----------------|
| Trust, data control, fast domain pack rollout | “Another ChatGPT wrapper” |
| On-prem / private cloud | Cloud-only SaaS from day one |

**Revenue by phase:**

| Phase | Revenue model |
|-------|---------------|
| A–B | Pilots, implementation, annual support |
| C | Hosted multi-tenant + subscription tiers |
| D | Partner channel + enterprise modules |

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

**Summary:** strong **technical MVP** and **platform foundation**. Missing **product maturity** for international B2B sales (UI polish, security narrative, enterprise features, packaged vertical).

---

## Phase A — International-ready product (0–4 months) ✅

**Goal:** credible demo and pilot for international buyers and integrators.

**Status:** Phase A deliverables are in the repository (`feature/phase-a-complete`).

### Product

| Item | Status | Artifact |
|------|--------|----------|
| **English-first UI** | ✅ | `webapp/`, `DEFAULT_LOCALE=en` |
| **Security brief** | ✅ | [SECURITY_BRIEF.md](./SECURITY_BRIEF.md) |
| **Pilot playbook** | ✅ | [PILOT_PLAYBOOK.md](./PILOT_PLAYBOOK.md) |
| **HR domain pack (EN)** | ✅ | [domain-packs/HR.md](./domain-packs/HR.md), `data/default/*_en.txt` |
| **Locale extensibility** | ✅ | [LOCALE_GUIDE.md](./LOCALE_GUIDE.md) |

### Engineering

| Item | Status | Artifact |
|------|--------|----------|
| Webapp i18n | ✅ | `/branding`, locale bundles |
| Expand eval | ✅ | `eval/rag_default_en_baseline.jsonl` (18 cases) |
| Smoke E2E in CI | ✅ | `smoke-api` job in CI |
| OpenAPI examples | ✅ | [API_EXAMPLES.md](./API_EXAMPLES.md) |

### GTM

- 2–3 pilot conversations (remote, English)
- 1 case study (anonymized OK)
- GitHub + `docs/en/` as primary entry point

### Success criteria

- Demo → pilot conversion ≥20%
- Pilot: ≥85% in-scope answers with citations
- Fresh deploy: **&lt;1 day**

**Revenue:** pilot **$8k–25k**.

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

### GTM

- **Annual license** (not card self-serve yet)
- Partner program v1: 1–2 integrators, rev share
- Trust center: security, architecture, subprocessors

### Success criteria

- 1–2 paid annual licenses
- Security questionnaire without per-client custom code
- Verify pass rate ≥75% on production eval
- Admin NPS ≥40

**Revenue:** annual license **$24k–80k** + support retainer.

---

## Phase C — Scalable platform & revenue (9–18 months)

**Goal:** repeatable revenue without 100% custom work per client.

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

### GTM

- Self-serve for SMB
- Enterprise sales for on-prem and large contracts

### Success criteria

- Positive MRR from hosted tier
- Healthy gross margin on hosted (LLM + infra)
- Annual contract churn &lt;5%
- New domain pack in **&lt;3 days**

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
4. **Domain pack = unit of sale** — platform is enabler, vertical is the offer.
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
MVP+i18n  →  EN product + HR pack → RBAC, audit, SSO,     → SaaS + billing +      → Partners +
             pilots + eval        dashboard               white-label             connectors
```

| Phase | Duration | Buyer | Revenue |
|-------|----------|-------|---------|
| **A** | 0–4 mo | HR / IT pilot sponsor | Pilot $8k–25k |
| **B** | 4–9 mo | CISO + HR + procurement | License $24k–80k/yr |
| **C** | 9–18 mo | SMB + enterprise | MRR + enterprise |
| **D** | 18+ mo | Partners | License + marketplace |

---

## Next 90 days

**Phase A product — done:**

1. ~~English-first webapp~~ ✅
2. ~~Security brief~~ ✅
3. ~~HR domain pack + demo script~~ ✅
4. ~~Eval + smoke CI~~ ✅

**Next — Phase B:**

5. Minimal audit log
6. RBAC
7. First paid pilot (GTM)

---

## Relation to old «Phase 3»

The previous list (Helm, SaaS, vision pack, audit, dashboard) is **split across Phases B–D** and tied to buyers and revenue.  
**Phase A** is the new prerequisite: international market does not open without it, even with good code.

---

See also: [ARCHITECTURE.md](./ARCHITECTURE.md), [DEPLOY.md](./DEPLOY.md), [Russian ROADMAP](../ru/ROADMAP.md).
