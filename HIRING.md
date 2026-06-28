# Hiring Portfolio — Grounded LLM

This document explains **why** Grounded LLM is built the way it is. Use it for technical interviews, cover letters, and architecture discussions.

**Author:** Kantemir Satibalov  
**Open to:** LLM Engineer · RAG Engineer · AI Platform Engineer · Backend Engineer (AI)  
**Locations:** Remote worldwide · Relocation to Canada, USA, Australia, New Zealand

---

## Elevator pitch (30 seconds)

> I built an open-source platform for deploying **document-grounded assistants** on-prem: cited answers, numeric verification, multi-tenant REST API, and eval-driven retrieval quality. Go handles auth and LLM orchestration; Python handles retrieval only. Teams can ship a new assistant from a template pack in days instead of rebuilding RAG from scratch.

---

## Role fit

| Job title | What this repo demonstrates |
|-----------|----------------------------|
| **LLM / RAG Engineer** | RAG pipeline, prompts, few-shot, verify layer, eval baselines |
| **AI Platform Engineer** | Multi-tenant API, streaming, OpenAPI, Docker deploy, observability |
| **Backend Engineer** | Go services, Postgres migrations, auth (Telegram + API keys), rate limits |
| **ML Engineer (applied)** | Embeddings, chunking, Chroma, retrieval metrics |

---

## Five design decisions (with trade-offs)

### 1. Split Go (orchestration) vs Python (retrieval)

**Decision:** LLM calls, sessions, verify, and admin live in Go. Python exposes only `/rag/context`.

**Why:** Single place for auth, logging, and API contracts; Python container stays focused on embeddings and Chroma. Easier to harden and scale the API tier independently.

**Trade-off:** Two runtimes to deploy. Mitigated by Docker Compose and clear HTTP boundary.

**Interview hook:** *“How would you scale this?”* — Scale Go replicas; Python RAG can be pooled or replaced with a managed vector service later.

---

### 2. Verify layer on numeric claims

**Decision:** After LLM response, extract numbers from the answer and check they appear in retrieved chunks. Failures surface a warning, not a silent hallucination.

**Why:** HR/policy use cases are sensitive to wrong days, amounts, and deadlines. Full semantic entailment is expensive; numbers are high-impact and testable.

**Trade-off:** Text hallucinations without numbers may pass. Roadmap: expand verify or add LLM-as-judge on staging eval.

**Interview hook:** Show `rag/verifier.py` and Go `finalizeRAGAnswer` — measurable quality gate.

---

### 3. Eval baselines before full LLM eval

**Decision:** JSONL suites test retrieval (`expect_contains`, `expect_context`, `expect_out_of_scope`) runnable against Python RAG alone.

**Why:** Fast CI feedback without burning LLM tokens; separates retrieval bugs from generation bugs.

**Trade-off:** Does not catch bad phrasing from the LLM. Roadmap: E2E eval with LLM on staging.

**Interview hook:** `python scripts/run_rag_eval.py --suite default_en` — 18 cases, EN HR demo.

---

### 4. Multi-tenant via header + data layout

**Decision:** `X-Tenant-ID` + `data/{tenant}/{domain}/` + Chroma metadata filter (`$and` on `tenant_id` and `domain_id`).

**Why:** B2B integrators need isolation without separate deployments per client.

**Trade-off:** Tenant ID in header is trust-boundary—must be set by authenticated gateway or API key policy.

**Interview hook:** Bug fix story — Chroma multi-filter required `$and` syntax; documented in vector store.

---

### 5. On-prem first, not cloud-only SaaS

**Decision:** Docker Compose, data residency narrative, security brief for IT reviewers.

**Why:** Enterprise buyers (CA, US, AU, NZ) block public ChatGPT for internal docs; they need cited answers in **their** VPC.

**Trade-off:** Harder self-serve growth than pure SaaS. Platform targets integrators and internal platform teams first.

---

## STAR stories (ready for interviews)

### Retrieval regression caught by eval

- **Situation:** Multi-tenant Chroma filter returned HTTP 500 for combined `domain_id` + `tenant_id`.
- **Task:** Restore similarity search without breaking single-field filters.
- **Action:** Aligned filter with `index_stats` pattern: `{"$and": [{"domain_id": ...}, {"tenant_id": ...}]}`; re-ran eval suite.
- **Result:** 18/18 EN baseline pass; documented for future domain packs.

### Production-minded admin surface

- **Situation:** KB editors need upload + reindex without SSH.
- **Task:** Secure admin API for non-developers.
- **Action:** Basic Auth on Go admin routes; `X-Admin-Secret` for Python reindex; filename whitelist and size limits.
- **Result:** Demo flow: upload TXT → reindex → new question answered with citation.

### Honest out-of-scope behavior

- **Situation:** Users ask questions outside the knowledge base.
- **Task:** Avoid confident hallucinations.
- **Action:** RAG returns empty/low context; prompts + eval edge cases (“CEO salary on the Moon”) expect out-of-scope.
- **Result:** Trust demo for security reviewers; KPI treats “not in KB” as correct.

---

## Tech stack map

| Area | Technologies |
|------|----------------|
| API / orchestration | Go, Gin, PostgreSQL |
| Retrieval | Python, Flask, LangChain, Chroma |
| Embeddings | `intfloat/multilingual-e5-small` (in-container) |
| LLM | OpenAI-compatible API (configurable endpoint) |
| Deploy | Docker Compose, nginx |
| Quality | pytest, go test, JSONL eval, GitHub Actions |

---

## Run locally in 10 minutes (demo for interviewers)

```bash
git clone https://github.com/kantik001/grounded-llm.git && cd grounded-llm
cp .env.example .env
# Set LLM_API_KEY; TELEGRAM_AUTH_DISABLED=true for browser demo

docker compose up -d --build
python scripts/reindex_rag.py
# Open http://localhost/?locale=en
# Ask: "How many paid vacation days do employees get?" → expect "28" + Sources block
```

```bash
pip install requests
python scripts/run_rag_eval.py --suite default_en   # retrieval metrics
make test                                          # unit tests
```

---

## What I would build next (shows senior judgment)

1. **Retrieval gate in CI** — spin up Python + Chroma in GitHub Actions and run `run_rag_eval.py`.
2. **Audit log** — upload, delete, reindex, admin login (enterprise blocker).
3. **Internal Python RAG auth** — shared secret or network policy; do not expose `:5000` on public host.
4. **Protect `/metrics`** — internal network or auth in production.
5. **Template CLI** — `grounded-llm init-pack hr` bundling config + eval scaffold.

---

## Links

- [README.md](README.md) — overview and quick start
- [PLATFORM_VISION.md](PLATFORM_VISION.md) — positioning
- [docs/en/ARCHITECTURE.md](docs/en/ARCHITECTURE.md)
- [docs/en/SECURITY_BRIEF.md](docs/en/SECURITY_BRIEF.md)
