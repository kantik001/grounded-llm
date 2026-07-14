# Ecosystem — Grounded standard vs adjacent projects

Grounded LLM is the **reference implementation** of an open standard for document-grounded assistants. Adjacent capabilities (agents, tool use, workflow graphs) belong in **separate projects** that may integrate via the public API — not inside the standard core.

See: [STANDARD_STRATEGY.md](./STANDARD_STRATEGY.md) · [PLATFORM_VISION.md](../../PLATFORM_VISION.md) · [RFC-0001](./rfcs/RFC-0001-grounded-compatible.md)

---

## What stays in this repository (standard core)

| Area | Why it belongs here |
|------|---------------------|
| **Grounded Spec v1** | Normative API + behavior contract |
| **Conformance CLI** | Testable «Grounded-compatible» label |
| **Retrieval quality** | Eval suites, benchmark, adversarial gates |
| **Citations + verify** | Core differentiators |
| **Vector / hybrid retrieval** | Measurable quality science (dense, BM25+RRF, adapters) |
| **Multi-tenant API** | Integrator surface |
| **Domain packs** | Template marketplace unit |
| **On-prem deploy** | Docker, Helm, Terraform reference |

Success metric: **new grounded assistant from template in &lt;3 days**, eval pass rate tracked on every release.

---

## What belongs in separate projects

| Capability | Separate project | Integration with Grounded |
|------------|------------------|---------------------------|
| **ReAct / tool-calling agents** | Agent runtime (future repo) | Calls `POST /api/v1/message` or `POST /rag/context` as a tool |
| **MCP gateway / registry** | MCP adapter service | Proxies tools; retrieval stays in Grounded |
| **Visual workflow builder** | Not planned | Out of scope per [STANDARD_STRATEGY.md](./STANDARD_STRATEGY.md) |
| **General chatbot (no KB)** | Not planned | Out of scope |
| **LLM inference serving (vLLM ops)** | Infra / MLOps stack | Grounded consumes OpenAI-compatible API |

**Rule:** if a feature requires arbitrary tool use or agent graphs, it does **not** enter Grounded Spec v1 without a new RFC and a major version bump.

---

## Standard-first roadmap (current)

### Horizon 1 — Reference implementation (now)

| Priority | Work | Pillar |
|----------|------|--------|
| ✅ | Hybrid retrieval (BM25 + dense + RRF) | 2, 3 |
| 🔜 | Hybrid modes documented in Grounded Spec §7 | 1 |
| ✅ | pgvector adapter (`VECTOR_STORE=pgvector`) | 2, 3 |
| 🔜 | Benchmark badge + `bench_report.py` in release flow | 2 |
| 🔜 | 4th domain pack with eval | 4 |

### Horizon 2 — Platform standard (6–18 months)

| Work | Pillar |
|------|--------|
| Retrieval mode conformance (vector vs hybrid) | 1, 2 |
| Connector ingest contract in spec | 1, 4 |
| Partner certification program | 5 |
| Alternate implementation passes conformance | 1, 5 |

### Horizon 3 — Industry standard (18+ months)

| Work |
|------|
| Public spec site (`grounded.dev`) |
| «Grounded-compatible» in RFP language |
| `grounded-bench` as cited benchmark |

---

## Agent project (planned, separate repo)

**Working name:** grounded-agents (or similar)

**Scope:**

- ReAct / LangGraph-style loops over **tools**
- One canonical tool: `grounded_retrieve` → Grounded `POST /rag/context`
- Optional: `grounded_chat` → `POST /api/v1/message`
- BDD step-library search, IDE plugins, bank QA scenarios — **consumer use cases**, not platform core

**Non-goals for agent repo:**

- Replacing Grounded retrieval or verify
- Forking domain packs or eval harness
- Claiming «Grounded-compatible» without running conformance against Grounded itself

**Hiring note:** agent work strengthens AI Engineer profile; standard work strengthens Platform / RAG Engineer profile. Both are valid; only the standard path lives in `grounded-llm`.

---

## Related

- [GROUNDED_SPEC_v1.md](./spec/GROUNDED_SPEC_v1.md)
- [BENCHMARK.md](./BENCHMARK.md)
- [GOVERNANCE.md](./GOVERNANCE.md)
- [COMPARISON.md](./COMPARISON.md)
