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
| **ReAct / tool-calling agents** | [grounded-agent](https://github.com/kantik001/grounded-agent) | Calls gRPC Retriever / `POST /rag/context` + MCP Gateway tools |
| **MCP gateway / registry** | [mcp-gateway](https://github.com/kantik001/mcp-gateway) | Proxies tools; retrieval stays in Grounded |
| **Token-level / remote verify** | [grounded-guardrails](https://github.com/kantik001/grounded-guardrails) | Optional `GUARDRAILS_MODE=remote` → gRPC `:50052` (`VerifyText`); default remains in-process Spec verify |
| **Verifiable-generation bench** | [grounded-bench](https://github.com/kantik001/grounded-bench) | Offline NVR / CP / HR / RR (dataset + predictions); retrieval gates stay in this repo `eval/` |
| **vLLM serving-path verify** | [grounded-vllm](https://github.com/kantik001/grounded-vllm) | OpenAI proxy `:8001` → vLLM `:8000` + guardrails `:50052`; set `LLM_BASE_URL` to the proxy |
| **Visual workflow builder** | Not planned | Out of scope per [STANDARD_STRATEGY.md](./STANDARD_STRATEGY.md) |
| **General chatbot (no KB)** | Not planned | Out of scope |

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
| ✅ | Spec-faithful row on [grounded-bench](https://github.com/kantik001/grounded-bench) (`make score-spec`) | 2 |
| 🔜 | 4th domain pack with eval | 4 |
| ✅ | [8-week pilot offer](./PILOT_OFFER.md) + day-0 script | 4 |

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
| `grounded-bench` as cited benchmark | **v0 shipped:** [grounded-bench](https://github.com/kantik001/grounded-bench) |

---

## Agent project (separate repo)

**Repo:** [github.com/kantik001/grounded-agent](https://github.com/kantik001/grounded-agent)

**Scope:**

- Go ReAct loop over **tools** (max 5 steps)
- Canonical tool: `retrieve[...]` → Grounded gRPC `Retriever` / HTTP `/rag/context`
- MCP tools via [mcp-gateway](https://github.com/kantik001/mcp-gateway): `call_tool[server.tool, {...}]`
- Redis session memory (`session:{id}`)
- Optional: compose `--profile full` with Grounded LLM GHCR images
- Offline demo: [docs/DEMO.md](https://github.com/kantik001/grounded-agent/blob/main/docs/DEMO.md) (`LLM_MODE=demo`)

**Non-goals for agent repo:**

- Replacing Grounded retrieval or verify
- Forking domain packs or eval harness
- Claiming «Grounded-compatible» without running conformance against Grounded itself

**Hiring note:** agent work strengthens AI Engineer / orchestration profile; standard work strengthens Platform / RAG Engineer profile. Both are valid; only the standard path lives in `grounded-llm`.

---

## Related

- [GROUNDED_SPEC_v1.md](./spec/GROUNDED_SPEC_v1.md)
- [BENCHMARK.md](./BENCHMARK.md)
- [GOVERNANCE.md](./GOVERNANCE.md)
- [COMPARISON.md](./COMPARISON.md)
