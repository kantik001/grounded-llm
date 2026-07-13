---
title: "I built a grounded RAG assistant for my father's research papers — then turned it into an open platform I want to become an industry standard"
description: Grounded LLM — Spec v1, conformance CLI, template packs, on-prem deploy. From one vertical proof to a checkable standard for cited document assistants.
tags: opensource, rag, ai, devops, standardization
series: Building Grounded LLM
canonical_url: https://github.com/kantik001/grounded-llm/blob/main/docs/en/blog/from-vertical-rag-to-open-standard.md
---

# I built a grounded RAG assistant for my father's research papers — then turned it into an open platform I want to become an industry standard

_Last month I published how I made ~500 horticulture papers queryable without hallucination ([passion project on DEV](https://dev.to/kantik001/my-father-wrote-the-papers-i-built-a-rag-assistant-so-growers-can-query-them-safely-1hi)). That vertical worked. This post is about what came next: **extracting the repeatable parts into an open platform — and publishing a spec other teams can conform to.**_

---

## The problem nobody demos on Twitter

Enterprise teams don't need another ChatGPT wrapper. They need assistants that:

1. Answer **only from internal documents** (policies, handbooks, runbooks).
2. **Cite sources** — filename + chunk — in every response.
3. **Refuse** when retrieval cannot support an answer.
4. Run **on infrastructure they control** (Docker, K8s, private cloud).
5. Ship with **measurable quality** — not "it looked fine in a notebook once."

I learned this building for scientific PDFs. The interesting work was never the chat bubble. It was making archives **answerable without lying** — and proving retrieval quality **before** burning LLM tokens.

That discipline became **[Grounded LLM](https://github.com/kantik001/grounded-llm)**.

---

## What I'm trying to standardize (and what I'm not)

I'm not building Dify, LangGraph, or a visual agent constructor.

I'm standardizing **one narrow class of systems**:

> **Document-grounded assistants** — internal Q&A with citations, numeric verification, and regression-tested retrieval.

The positioning line:

> *Open standard for document-grounded assistants with citations, numeric verify, and measurable retrieval quality — deployable on your infrastructure.*

**Non-goals (by design):**

- Arbitrary tool/agent graphs
- General chat without a knowledge base
- Cloud-only lock-in
- Feature parity with Glean or Microsoft Copilot SaaS

We compete on **trust + reproducible quality + conformance** — not feature count.

---

## Five pillars of the standard

| Pillar | What it means | Today in the repo |
|--------|---------------|-------------------|
| **1. Spec & conformance** | Published rules + tests anyone can run | [Grounded Spec v1](https://github.com/kantik001/grounded-llm/blob/main/docs/en/spec/GROUNDED_SPEC_v1.md), `python -m conformance` |
| **2. Quality science** | Numbers, not demos | JSONL eval suites, **retrieval gate in CI**, adversarial pack |
| **3. Reference deploy** | Reproducible install | Docker, Helm, Terraform (AWS/GCP/Azure) |
| **4. Template marketplace** | Growth without forking core | HR, IT Support, Legal FAQ packs + registry |
| **5. Governance** | Standard outlives one author | [RFC process](https://github.com/kantik001/grounded-llm/blob/main/docs/en/RFC.md), [RFC-0001 Grounded-compatible](https://github.com/kantik001/grounded-llm/blob/main/docs/en/rfcs/RFC-0001-grounded-compatible.md) |

Horizon 1 success metric: **any engineer runs conformance on a fresh deploy in under 15 minutes.**

---

## What the platform is today (not a slide deck)

After 11 delivery phases merged to `main`, this is a **working reference implementation**, not a manifesto.

### Architecture

```
Clients (Web / API / Telegram)
        ↓
Go server — auth, sessions, LLM, verify, admin, quotas
        ↓ POST /rag/context
Python RAG — embeddings, Chroma/Qdrant, hybrid rerank
        ↓
data/{tenant}/{domain}/  +  Postgres
```

**Split on purpose:** Go owns trust boundaries and orchestration; Python owns retrieval only.

### What ships out of the box

| Capability | Why it matters |
|------------|----------------|
| **Citations in every answer** | Audit trail for HR/legal |
| **Numeric verify layer** | Dosages, vacation days, SLA numbers must match retrieved context |
| **Retrieval eval gate in CI** | Catches silent RAG regressions on every PR |
| **Multi-tenant API** | `X-Tenant-ID`, API keys, OpenAPI v1 |
| **Enterprise admin** | RBAC, OIDC SSO, audit log, async reindex, analytics |
| **Template packs** | `python scripts/init_pack.py install hr` |
| **Ingest connectors** | SharePoint, Google Drive, Confluence |
| **Conformance CLI** | Offline spec check + live HTTP check against any deployment |
| **Embeddable widget** | Intranet embed, not only Telegram |

### Try conformance in 3 commands

```bash
git clone https://github.com/kantik001/grounded-llm.git
cd grounded-llm
pip install -r conformance/requirements.txt
python -m conformance spec          # offline OpenAPI contract
# after docker compose up:
python -m conformance check --url http://localhost:8080
```

If your product is **Grounded-compatible**, these tests should pass without forking my codebase.

---

## From one vertical to a platform (the story arc)

| | Horticulture proof | Grounded LLM platform |
|--|-------------------|----------------------|
| **Repo** | [grounded_horticulture_en](https://github.com/kantik001/grounded_horticulture_en) | [grounded-llm](https://github.com/kantik001/grounded-llm) |
| **Domain** | Apple rootstocks, disease IDs | Any internal documents |
| **Retrieval** | Heavy hybrid tuning (BM25 + vectors + RRF) | Pluggable; eval gate is the contract |
| **Deliverable** | Demo corpus + passion story | Spec + packs + enterprise deploy |

The horticulture project answered: *"Can we make scientific PDFs queryable safely?"*

Grounded LLM answers: *"Can we ship the **next** assistant in days without rebuilding auth, verify, eval, and deploy?"*

---

## Media: what to show (GIFs & screenshots)

> **For the published DEV post:** replace each `📸` block below with an asset.  
> Capture guide: [DEVTO_PLATFORM_ARTICLE_MEDIA.md](./DEVTO_PLATFORM_ARTICLE_MEDIA.md)

### Cover image (required)

📸 **Suggested cover:** split screen — left: Spec v1 doc + conformance terminal green; right: chat UI with visible **citation chips**. Text overlay: *"Grounded LLM — open standard for cited assistants"*

### GIF 1 — Chat with citations (hero demo)

📸 Record 30–45s: HR or IT Support pack, 2–3 questions, show **source filenames** in the answer. English UI preferred for international audience.

Example questions:
- *How many paid vacation days do employees get?*
- *What is the password reset SLA?*

### GIF 2 — Conformance CLI (the "standard" moment)

📸 Terminal recording:

```bash
python -m conformance spec
python -m conformance check --url http://localhost:8080
```

Green output = the punchline. This is what makes the post different from another RAG tutorial.

### GIF 3 — CI retrieval gate (optional, for engineers)

📸 GitHub Actions `eval-retrieval-gate` job green on a PR — proves quality is enforced, not demoed.

### Screenshot 4 — Template packs

📸 Terminal: `python scripts/init_pack.py list` + registry table from `packs/registry.yaml`.

### Screenshot 5 — Admin panel (enterprise credibility)

📸 `admin.html`: upload document → reindex → index stats. Shows it's deployable, not a weekend script.

---

## Where this sits in the industry (including Google)

Big tech is solving **adjacent** problems:

| Product / area | Focus | Grounded LLM difference |
|----------------|-------|-------------------------|
| **NotebookLM** | Research / consumer grounding on uploads | We target **enterprise on-prem**, API contract, CI gates |
| **Vertex AI Search** | Managed cloud retrieval | We target **self-hosted**, MIT core, conformance badge |
| **Gemini + Workspace** | SaaS copilot inside Google | We target **any LLM endpoint**, any infra |

I'm not competing with Google on consumer UX. I'm saying: **when procurement asks "is your internal assistant grounded and testable?" — there should be a published spec and CLI answer, not a vendor slide.**

If you work on **enterprise RAG, OSS standards, or ML platform conformance** — I'd genuinely value feedback on [RFC-0001](https://github.com/kantik001/grounded-llm/blob/main/docs/en/rfcs/RFC-0001-grounded-compatible.md).

> **Note on @mentions:** I don't tag individuals cold on social posts. I use tags `#opensource` `#rag` `#ai` `#standardization` and engage in comments. For Google folks specifically: the relevant conversation is **enterprise grounding + eval contracts**, not "please promote my repo."

---

## What happens next (my honest roadmap)

Delivery phases **1–11 are complete** in code. The next work is strategic, not "Phase 12 features for the sake of it":

### Track 1 — Launch the standard (now)

- [ ] Public repository + tag **`v0.3.0`**
- [ ] GitHub Pages site live (spec + packs index)
- [ ] This article + Show HN / r/selfhosted
- [ ] External teams run `conformance check` and report gaps

### Track 2 — Prove adoption (6–12 months)

- [ ] 3+ production deployments **not** maintained by me
- [ ] 1 alternate implementation passes conformance (partial is fine)
- [ ] Partner certification pilot ([docs](https://github.com/kantik001/grounded-llm/blob/main/docs/en/PARTNER_CERTIFICATION.md))

### Track 3 — Optional hosted path (only if demand)

- Signup + Stripe already scaffolded — **not** the main story until someone pays for hosted
- Enterprise SAML, trust center refresh for pilots

**Horizon 3 north star:** *"Grounded-compatible"* shows up in RFPs and security questionnaires — like "OpenTelemetry-compatible" does for observability today.

---

## Call to action

1. **Star / watch** the repo if enterprise grounding interests you: [github.com/kantik001/grounded-llm](https://github.com/kantik001/grounded-llm)
2. **Run conformance** on your deploy and open an issue if something should be in Spec v2
3. **Contribute an eval case** when you fix a retrieval bug — see [GOOD_FIRST_ISSUES.md](https://github.com/kantik001/grounded-llm/blob/main/GOOD_FIRST_ISSUES.md)
4. **Building in horticulture or another vertical?** The passion repo is still the deep retrieval story; this platform is the generalization

---

## Summary

| Question | Answer |
|----------|--------|
| What is it? | Open platform + Spec v1 for cited, verified document assistants |
| What is it not? | Agent builder, ChatGPT clone, Glean competitor |
| Why now? | Vertical proof worked; standard + conformance is the multiplier |
| What's next? | Public launch, external conformance adopters, RFP-grade positioning |

I started with my father's papers. I want to end with a **checkable standard** any team can implement — and prove — on their own infrastructure.

---

**Series:** [Passion project — horticulture RAG](https://dev.to/kantik001/my-father-wrote-the-papers-i-built-a-rag-assistant-so-growers-can-query-them-safely-1hi) · **Repo:** [grounded-llm](https://github.com/kantik001/grounded-llm) · **Spec:** [GROUNDED_SPEC_v1.md](https://github.com/kantik001/grounded-llm/blob/main/docs/en/spec/GROUNDED_SPEC_v1.md)

*Disclaimer: Grounded LLM is MIT-licensed reference software. Assistant output is informational; compliance and field decisions require human review.*
