# Platform Vision — Grounded LLM

## One line

> **Open platform to deploy cited, verified document assistants in days — templates, API, on-prem.**

---

## What we are

Grounded LLM is a **narrow, opinionated platform** for one class of business AI:

**Document-grounded assistants** — internal Q&A over policies, handbooks, and knowledge bases, with:

- Answers **only from uploaded documents**
- **Citations** (filename + chunk) in every response
- **Verification** of numeric claims against retrieved context
- **Eval baselines** to measure retrieval regression
- **Multi-tenant API** for integrators
- **On-prem / private cloud** deploy via Docker

**Template pack model:** platform core stays stable; each use case ships as a **domain pack** (config + documents + eval + locale bundles). Reference template: [HR policy assistant](docs/en/domain-packs/HR.md).

**Time to first assistant (engineer-led):** deploy platform → add template → upload docs → reindex → run eval. Target: **2–5 days** with documents ready.

---

## What we are not

| We are not | Why |
|------------|-----|
| **Universal AI constructor** (Dify, Flowise) | No visual workflow builder, no arbitrary agent graphs |
| **LLM framework** (LangChain, LlamaIndex) | We ship a deployable product, not a library |
| **Vertical SaaS** (Glean, etc.) | Open core + self-host; you own the stack |
| **General chatbot** | No knowledge base → honest “not found”, not improvisation |

We win on **trust, deployability, and measured RAG quality**—not on feature count.

---

## Platform vs template pack

```
┌─────────────────────────────────────────────────────────┐
│  Platform core (this repo)                              │
│  Go orchestration · Python retrieval · verify · admin   │
│  Multi-tenant API · eval harness · Docker               │
└───────────────────────────┬─────────────────────────────┘
                            │
         ┌──────────────────┼──────────────────┐
         ▼                  ▼                  ▼
    HR template       IT support template   Legal FAQ template
    config + data     config + data         config + data
    + eval            + eval                + eval
```

| | Platform core | Template pack |
|---|---------------|-----------------|
| Changes | Rarely | Per customer / use case |
| Paths | `server/`, `api/`, `rag/` | `config/`, `data/{tenant}/{domain}/`, `eval/` |
| Owner | Platform maintainers | Implementers, integrators |

---

## Differentiators

| Capability | Grounded LLM | Typical tutorial RAG |
|------------|--------------|----------------------|
| Citations in UI | ✅ | Sometimes |
| Verify numbers vs context | ✅ | Rare |
| Eval baseline + CI | ✅ | Rare |
| Multi-tenant isolation | ✅ | Rare |
| OpenAPI + API keys | ✅ | Varies |
| On-prem Docker | ✅ First-class | Often cloud-only |
| Security brief for IT | ✅ | Rare |

---

## Comparison (honest)

| | Grounded LLM | LangChain DIY | No-code builders | Enterprise SaaS |
|---|--------------|---------------|------------------|-----------------|
| Time to prod | Days (with template) | Weeks–months | Hours (simple) | Weeks (procurement) |
| Data control | Full (on-prem) | You build it | Often cloud | Vendor cloud |
| Cited + verify | Built-in | You build it | Varies | Varies |
| Customization | Code + config | Full | Limited | Limited |
| Best for | Teams shipping **trusted** internal assistants | Research, custom stacks | Quick prototypes | Large orgs with budget |

---

## Roadmap (platform-focused)

### Done ✅

- Platform core: RAG pipeline, citations, verify, admin, migrations
- Multi-tenant API, streaming, OpenAPI, observability
- English-first UI, locale packs, HR reference template
- Eval baselines (EN + RU), smoke API in CI, **retrieval eval gate in CI**

### Next — enterprise hardening (Phase B)

- RBAC, audit log, SSO (OIDC/SAML)
- Async reindex with job status
- Retrieval improvements (reranker) measured via eval
- Helm chart, backup/restore runbooks

### Later — scale (Phase C–D)

- Template catalog / marketplace (HR, IT, legal FAQ)
- Connectors: SharePoint, Google Drive, Confluence
- Optional hosted multi-tenant tier
- Open core: MIT platform + commercial enterprise modules

Success metric for the platform: **new grounded assistant from template in &lt;3 days**, with eval pass rate tracked on every release.

---

## Who this is for

- **Platform / backend engineers** building internal AI for enterprises
- **Integrators** deploying on-prem assistants for clients
- **Teams** that need citations and audit-friendly answers—not a black-box chatbot

---

## Related docs

- [HIRING.md](HIRING.md) — design decisions for technical interviews
- [docs/en/ARCHITECTURE.md](docs/en/ARCHITECTURE.md)
- [docs/en/ROADMAP.md](docs/en/ROADMAP.md)
- [docs/en/SECURITY_BRIEF.md](docs/en/SECURITY_BRIEF.md)
