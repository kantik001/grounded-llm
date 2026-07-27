# Optimized Career Plan — Verifiable AI Infrastructure
## Goal: AI Infrastructure Engineer track → high TC ($1M trajectory)
## Version: 1.1 | Date: 2026-07-27
## Supersedes execution priority of: VERIFIABLE_AI_PLATFORM_INSTRUCTION_v1.md

---

## 0. Reality check on “$1M as fast as possible”

| Path | Realistic timeline | What actually moves the needle |
|------|--------------------|--------------------------------|
| **Employed AI Infra (FAANG / frontier / GPU cloud)** | 3–7 years to **$400–800k TC**; **$1M+** = Staff/Principal + strong equity year | Depth in inference serving, merged OSS, measurable perf, production ownership |
| **Early eng at well-funded inference/infra startup** | 2–5 years to **$1M+** via equity (high variance) | Same technical signal + join before Series B if possible |
| **Founder / indie products** | High variance | Products alone rarely print $1M without distribution or acquisition |

**Implication:** Optimize for **hireability into top AI Infra roles** and **optional startup optionality**, not for building 8 repos. Stars and HN are marketing; **merged PRs + production depth + public benches with numbers** are the credential.

Qwen’s stack is directionally correct. Weight it by market demand:

| Skill | Hiring weight (AI Infra) | How to prove it |
|-------|--------------------------|-----------------|
| **Python + vLLM / serving APIs** | Critical | Plugin, PR, latency/throughput numbers |
| **C++/CUDA (or deep perf in serving path)** | High for senior+ | Real hot-path work, nsight/profile, not toy kernels |
| **Go (orchestration, gRPC, control plane)** | High | You already have this — deepen, don’t restart |
| **Rust (sandbox / streaming / safety)** | Differentiator, not gate | One sharp artifact (WASM tools or token guardrails) |

Do **not** chase equal depth in all four at once. Sequence: **Go (keep) → Python/vLLM (gain) → Rust (one wedge) → CUDA (only when you have a hot path)**.

---

## 1. Keep polyrepo. Do not merge into a monorepo.

Existing assets stay independent (history, CI, GHCR, docs site):

- `grounded-llm` — RAG platform + Spec + conformance (credibility core)
- `mcp-gateway` — tool control plane
- `grounded-agent` — orchestration demo

New work = **new focused repos** that **plug into** the above. Optional umbrella: a thin `verifiable-ai` org/README + compose that wires them — not a git merge.

---

## 2. Product portfolio optimized for career (not completeness)

Build **three** new artifacts max in the next 12 months. Everything else is backlog.

### Tier A — Must ship (career-critical)

#### A1. `grounded-guardrails` (Rust core + Go gRPC + thin Python client)
**Skills:** Rust (zero-copy / streaming), Go (gRPC), Python (client + later vLLM hook)  
**Story:** “Token-level verification for grounded generation — block bad numbers/PII mid-stream.”  
**Integrate:** grounded-llm verify path + grounded-agent final answer check.  
**Port:** **not** `:50051` (Retriever owns it) → e.g. `:50052`.  
**Scope cut vs v1 instruction:**
- Ship: ring buffer, numeric extract/verify, PII regex, gRPC `VerifyText` + `VerifyStream`, benches, Docker
- Defer: ONNX toxicity (commodity; weak differentiator), triple reimplementation of numeric verify
- One algorithm: Rust = source of truth; Python/Go call via gRPC or FFI — no three divergent copies

**Acceptance for resume:** p99 overhead published; wired into grounded-llm in a release note.

#### A2. vLLM contribution + thin `grounded-vllm` adapter
**Skills:** Python (critical), then C++/CUDA only if a real bottleneck appears  
**Story:** “Upstream contribution + callback that runs grounded verify in the serving path.”  
**Sequence:**
1. Read vLLM plugin/callback surfaces; ship a **working adapter** that calls guardrails over gRPC/HTTP
2. Open **useful** PRs (docs, bugfix, small feature maintainers want) — aim for **merged**, not flashy
3. Only after profiling: CUDA/kernels for a measured hotspot

**This is the highest ROI item for “AI Infrastructure Engineer” titles.** Portfolio RAG without serving-path work reads as App/Platform; vLLM reads as Infra.

#### A3. `grounded-bench` (Python dataset + Go or Python runner + static leaderboard)
**Skills:** Python eval metrics, Go runner optional  
**Story:** “Public benchmark for verifiable / grounded generation (NVR, citation precision, hallucination, refusal).”  
**Leverage:** expand existing `grounded-llm/eval` + conformance — don’t rebuild distributed workers on day one.  
**Cut:** 50-worker Rust distributed eval → later. Start: single-node CLI, 250–1000 cases, reproducible seed, published numbers for your stack.

### Tier B — Strong differentiator (after A1–A3 have public artifacts)

#### B1. WASM sandboxed tools **inside** `mcp-gateway` (not a second gateway)
**Skills:** Rust guests + Go host (wazero)  
**Story:** “MCP tools with capability-based isolation — same HTTP API, safer execution.”  
**Cut:** drop-in twin gateway; extend mcp-gateway with a `runtime: wasm` server type.  
**Note:** use **WASI / C ABI** for wazero — not `wasm-bindgen` (browser-oriented).

### Tier C — Explicitly defer (high effort, low early ROI)

| Item | Why defer |
|------|-----------|
| Full monorepo `verifiable-ai-platform` | Breaks existing distribution; no hire signal |
| `grounded-eval` distributed Rust workers | Commodity; start from bench/conformance |
| CUDA `pii_detect` / `numeric_verify` kernels | CPU already μs; GPU needs proven vLLM hot path |
| ONNX toxicity as Phase 1 centerpiece | Crowded space; weak unique claim |
| HN frontpage as success criterion | Uncontrollable; replace with “published bench + 1 merged upstream PR” |

---

## 3. 12-month sequence (career-weighted)

```text
Months 1–2   Guardrails Rust MVP + benches + gRPC (:50052)
Months 2–3   Wire into grounded-llm (+ agent); release notes with numbers
Months 3–5   grounded-bench v0 (250–1000 cases) + public leaderboard page
Months 4–6   vLLM adapter + first upstream PR(s)
Months 6–8   WASM runtime as mcp-gateway backend (calculator + sql-sandbox)
Months 8–12  Only if profiled: CUDA/kernels OR deeper vLLM internals; talks/articles
             Parallel always: LinkedIn narrative, applications, network
```

**Parallel career track (non-optional if $1M is the goal):**
- Target companies: inference platforms, GPU clouds, frontier labs infra, AI platform teams
- Weekly: 2–3 tailored applications / warm intros once A1+A2 have links
- One English case study: architecture + flamegraph/bench table + “what I’d do next”
- Track TC bands (levels.fyi) for Staff AI Infra / Inference Engineer — aim role ladder, not job title poetry

---

## 4. Skill → artifact map (what recruiters can click)

| Skill Qwen listed | Artifact that proves it | When |
|-------------------|-------------------------|------|
| Python / ML serving / eval | vLLM adapter + grounded-bench metrics | M3–6 |
| C++/CUDA | Optional kernels **after** profile; or contribute to existing CUDA path in vLLM | M8+ |
| Go | grounded-llm, mcp-gateway, guardrails gRPC, agent | Already + deepen |
| Rust | Guardrails token path; later WASM guests | M1–2, M6–8 |

---

## 5. Success criteria (controllable vs vanity)

### Controllable (optimize for these)

```yaml
technical:
  guardrails_wired_into_grounded_llm: true
  published_benches_with_hardware_notes: true
  grounded_bench_public_v0: true
  vllm_adapter_demo: true
  upstream_prs_merged: ">= 1"   # stretch: 2
  wasm_tools_in_mcp_gateway: true  # after A-tier

career:
  english_case_study: 1
  target_company_conversations: ">= 15"
  onsite_or_final_rounds_infra: ">= 3"
  offer_track: "AI Infra / Inference / Platform Eng with equity upside"
```

### Vanity (nice, not gates)

```yaml
stars_total: 1000+          # do not block shipping on this
hn_frontpage: 2+            # uncontrollable
```

---

## 6. Rules for Cursor agents (unchanged spirit, tighter scope)

1. One acceptance-gated step at a time.
2. Prefer **integrating into existing repos** over new umbrellas.
3. No CUDA until a benchmark shows CPU path is the bottleneck **in a serving integration**.
4. No second HTTP gateway — extend mcp-gateway.
5. English for code, commits, README.
6. Commit only when the user asks (repo rule) unless user says otherwise for this program.
7. If a step fails acceptance — stop, report, propose fix.

---

## 7. Immediate next step

**Status (2026-07-28):** Tier A shipped — `grounded-guardrails` (wired), `grounded-bench` v0, `grounded-vllm` v0 + overhead table + upstream docs PR.

When continuing:

1. Land / iterate [vLLM docs PR #50051](https://github.com/vllm-project/vllm/pull/50051); keep engaging [RFC #43999](https://github.com/vllm-project/vllm/issues/43999).
2. Optional: English case study (arch + [OVERHEAD.md](https://github.com/kantik001/grounded-vllm/blob/main/OVERHEAD.md) + bench numbers).
3. Tier B: WASM sandboxed tools **inside** `mcp-gateway` (wazero / WASI).
4. Keep `docs/plans/VERIFIABLE_AI_PLATFORM_INSTRUCTION_v1.md` as detailed backlog; **this file wins on priority**.

---

## 8. One-line strategy

> Ship verifiable inference plumbing that touches the vLLM serving path, prove it with a public bench, and use Go/Rust only where they create isolation or control-plane leverage — then convert that signal into Staff-track AI Infra interviews with equity upside.
