# Standard strategy — Grounded LLM

How we move from **reference implementation** to **industry standard** for verified, cited, on-prem document assistants.

See also: [PHASE_5.md](./PHASE_5.md) · [PLATFORM_VISION.md](../../PLATFORM_VISION.md)

---

## Positioning (one line)

> Open standard for **document-grounded assistants** with citations, numeric verify, and measurable retrieval quality — deployable on your infrastructure.

We do **not** compete with Dify/LangGraph on workflow builders. We compete on **trust + reproducible quality + conformance**.

---

## Five pillars

| # | Pillar | Meaning | Phase 4 | Phase 5+ |
|---|--------|---------|---------|----------|
| **1** | **Spec & conformance** | Published rules + tests anyone can run | Policy, offline/HTTP tests | Spec v1, `grounded-conformance` CLI |
| **2** | **Quality science** | Numbers, not demos | Adversarial JSONL, eval gate | Public bench, leaderboard badge |
| **3** | **Reference deploy** | Reproducible install | Docker, Helm, compatibility matrix | Terraform, vector adapters (Phase 6) |
| **4** | **Template marketplace** | Growth without forking core | HR, IT packs | Legal/compliance packs, pack registry (Phase 6) |
| **5** | **Governance & community** | Standard outlives one author | CONTRIBUTING, GOOD_FIRST_ISSUES | RFC process, steering, certification (Phase 7) |

Each phase PR should state: **which pillar(s)** and **which horizon** it advances.

---

## Three horizons

### Horizon 1 — Reference implementation (0–6 months after Phase 4)

**Goal:** Any engineer deploys an assistant with green conformance and eval pass.

| Work | Pillar |
|------|--------|
| Grounded Spec v1 + conformance CLI | 1 |
| Public benchmark / release badges | 2 |
| RFC-0001 «Grounded-compatible» | 5 |
| Tag v0.3.x releases with conformance | 1 |

**Success:** External repo runs `python -m conformance check` in &lt;15 min.

### Horizon 2 — Platform standard (6–18 months)

**Goal:** Integrators and vendors build on the contract.

| Work | Pillar |
|------|--------|
| Vector store adapters, reranker/hybrid | 2, 3 |
| Template catalog (legal, compliance) | 4 |
| Connectors (SharePoint, Drive) | 4 |
| Optional hosted tier + billing | 3 (path B) |
| Embeddable widget | 4 |

**Success:** 3+ production deployments not maintained by core team; 1 partial alternate implementation passes conformance.

### Horizon 3 — Industry standard (18+ months)

**Goal:** «Grounded-compatible» appears in RFPs and analyst material.

| Work | Pillar |
|------|--------|
| grounded.dev spec site | 1, 5 |
| grounded-bench as cited benchmark | 2 |
| Partner certification program | 5 |
| Enterprise module (SAML, DLP) — path B | 3 |

---

## Two business paths (compatible)

| | **A — Open standard** | **B — Product company** |
|--|------------------------|-------------------------|
| Focus | Rules, fame, ecosystem | Revenue, convenience |
| Money from | Support, training, packs, certification | Subscription, cloud, enterprise module |
| Code | Mostly MIT forever | MIT core + paid enterprise/hosted |
| Order | **Now → Horizon 1–2** | **After Horizon 1** (optional) |

**Recommended:** A first (Phase 5–6), add B when there is demand for hosted/SLA (Phase 7+).

---

## Explicit non-goals (all horizons)

- Visual agent/workflow builder
- General chatbot without knowledge base
- Country-specific compliance packages without a client
- Feature parity with Glean/Copilot SaaS

---

## Metrics that prove progress

| Metric | Horizon 1 target |
|--------|------------------|
| Conformance CLI pass (mock deploy) | 100% on reference impl |
| Retrieval eval (all suites) | 100% in CI |
| Adversarial retrieval suite | 100% in CI |
| Documented domain packs | ≥4 with eval |
| External contributors | ≥5 merged PRs |
| RFCs accepted | ≥1 (RFC-0001) |

---

## Related

- [ROADMAP.md](./ROADMAP.md)
- [API_DEPRECATION_POLICY.md](./API_DEPRECATION_POLICY.md)
- [BENCHMARK.md](./BENCHMARK.md)
- [RFC.md](./RFC.md)
