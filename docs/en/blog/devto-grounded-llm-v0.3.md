---
title: "Grounded LLM v0.3.0 shipped — and I finally drew an architecture diagram I'm proud of"
tags: ai, opensource, rag, architecture, docker, go
canonical_url: https://dev.to/kantik001/grounded-llm-v03-shipped-local-llms-redis-grpc-and-99-eval-gate
cover_image: docs/assets/architecture.png
series: Grounded LLM
---

> **Publishing notes (remove before posting):**
> - **Cover image:** upload `docs/assets/architecture.png` — this is the hero asset.
> - **Series:** link to [Part 1 — Building an open standard for grounded document assistants](https://dev.to/kantik001/building-an-open-standard-for-grounded-document-assistants-2h6e).
> - **Suggested slug:** `grounded-llm-v03-shipped-local-llms-redis-grpc-and-99-eval-gate`

---

Three weeks ago I published [Building an open standard for grounded document assistants](https://dev.to/kantik001/building-an-open-standard-for-grounded-document-assistants-2h6e).

Back then, Grounded LLM was **v0.1.0** — a credible reference implementation with a spec, conformance CLI, hybrid retrieval, and **89** eval cases in CI.

Honest take: it was the right *story*. But it still felt like a very good **v1 of the idea**.

Since then I shipped **v0.2.0** (production hardening) and, this week, **[v0.3.0](https://github.com/kantik001/grounded-llm/releases/tag/v0.3.0)**.

This post is not a changelog dump. It's what changed in my head while building — and why I think enterprise RAG needs a diagram like this, not another demo GIF.

---

## The moment I knew v0.3 was different

I was testing the same HR question for the fifth time:

> *"How many paid vacation days do employees get?"*

First run: full path — retrieval, LLM, verify, citations. ~4 seconds on local Ollama.

Second run: **instant**. Header: `X-Cache: HIT`.

Same answer. Same citations. No tokens burned.

That's when it stopped being "a RAG repo with good docs" and started feeling like **infrastructure**.

Not because caching is novel. Because the whole system — retrieval, verify, cache keys, metrics, agent surface — finally snapped into one picture.

So I drew it:

![Grounded LLM Architecture v0.3](https://github.com/kantik001/grounded-llm/raw/main/docs/assets/architecture.png)

*Clients → Go orchestration → Python RAG → Storage / Redis / LLM — with a separate agent path over gRPC.*

This is the diagram I wanted twelve weeks ago when I was still explaining "grounded" to people who'd only seen ChatGPT wrappers.

---

## What Grounded LLM is (30-second version)

**Grounded LLM** is an open platform + [Grounded Spec v1](https://github.com/kantik001/grounded-llm/blob/main/docs/en/spec/GROUNDED_SPEC_v1.md) for **document-grounded assistants**:

- answers **only from your knowledge base**
- **cites sources** in every response
- **refuses** when retrieval can't support an answer
- **verifies numbers** against retrieved context (±0.01)
- runs **on infrastructure you control**
- ships with **measurable retrieval quality** — not "looked fine in a notebook"

Positioning line:

> *Open standard for document-grounded assistants with citations, numeric verify, and measurable retrieval quality — deployable on your infrastructure.*

**Non-goals:** arbitrary agent graphs, general chat without KB, cloud lock-in, Glean feature parity.

We compete on **trust + reproducible quality + conformance** — not widget count.

Landing: [kantik001.github.io/grounded-llm](https://kantik001.github.io/grounded-llm/)

---

## v0.1 → v0.3 in one table

| | **v0.1** (first DEV post) | **v0.3.0** (today) |
|---|---|---|
| **Story** | "Here's an open standard" | "Here's a deployable platform" |
| **Retrieval eval** | 89 cases | **99** cases (+ adversarial near-miss) |
| **LLM** | Cloud API | Cloud **or** Ollama **or** vLLM — env switch only |
| **Caching** | — | Redis: embeddings + semantic response cache |
| **Agent surface** | HTTP only | **gRPC Retriever** on `:50051` |
| **Observability** | Basic | Prometheus: `llm_tokens_*`, TTFT, latency, cache hits |
| **Images** | — | GHCR `:0.3.0` for server, python, webapp |
| **Hardening** | — | Prod fail-fast, Trivy, backup smoke in CI, Helm probes |

The spec didn't change its soul. The **reference implementation grew teeth**.

---

## Walk through the diagram (the fun part)

### 1. Clients — four doors, one trust boundary

- **Web chat** — reference UI
- **Python SDK / REST** — integrators
- **Telegram Mini App** — optional field channel
- **Agents (gRPC)** — the new door in v0.3

Same platform. Different entry points. Go still owns auth, sessions, and policy.

### 2. Go server `:8080` — orchestration, not retrieval

Go does what enterprises actually audit:

| Module | Job |
|--------|-----|
| **Auth** | API keys, Telegram WebApp, OIDC |
| **REST API v1** | OpenAPI contract, multi-tenant |
| **LLM orchestration** | Provider routing, streaming, cache lookup |
| **Numeric verify** | Reject answers with numbers not in context |
| **Admin** | KB upload, reindex jobs, RBAC |

Python does retrieval. Go does trust. **On purpose.**

### 3. Python RAG — hybrid retrieval + agent contract

```
HTTP  :5000  →  /rag/context     (chat path)
gRPC  :50051 →  Retriever/Retrieve (agent path)
```

Under the hood:

- **Hybrid BM25 + dense + RRF**
- Vector backends: **Chroma / Qdrant / pgvector**
- Optional reranker (keyword or cross-encoder)

Agents don't need to reimplement your chunking strategy. They call the same retrieval the chat UI uses — with a stable protobuf contract (`grounded.rag.v1`).

### 4. Redis `:6379` — the silent hero

Two caches, two lifetimes, two owners:

| Cache | Key pattern | TTL | Who writes | Signal |
|-------|-------------|-----|------------|--------|
| **Embeddings** | `embedding:{md5}:{model}` | 1h | Python | `rag_embedding_cache_hit_total` |
| **LLM responses** | `response:{md5}:{model}` | 24h | Go | HTTP `X-Cache: HIT\|MISS` |

Why this matters for HR/legal assistants:

- Repeat policy questions are **free** after the first hit
- Re-indexing doesn't silently change cached answers (keys include domain/tenant/model)
- You can **prove** cache behavior in HTTP traces — procurement teams love receipts

### 5. LLM — one interface, three realities

```bash
LLM_PROVIDER=openai   # OpenRouter / any OpenAI-compatible cloud
LLM_PROVIDER=ollama   # docker compose --profile ollama
LLM_PROVIDER=vllm     # docker compose --profile vllm  (NVIDIA)
```

No code forks. No "enterprise edition" switch. **Env vars + Compose profiles.**

That's the on-prem story in one sentence: *your documents, your GPUs, your audit trail.*

---

## The quality gate nobody demos on Twitter

Grounded LLM ships **99 retrieval eval cases** across EN/RU/IT/Legal/Adversarial suites.

CI job `eval-retrieval-gate` runs on **every PR**:

```bash
python scripts/run_rag_eval.py --suite all
```

Example case:

```json
{"question": "How many paid vacation days?", "expect_contains": ["28"], "expect_context": ["28"]}
```

Hits Python RAG directly — **no LLM tokens burned** on the hot path.

| Suite | Cases | What it catches |
|-------|-------|-----------------|
| EN HR baseline | 21 | Paraphrase drift |
| Adversarial | 30 | Wrong numbers, cross-domain, injection |
| IT / Legal templates | 29 | Pack-specific retrieval |
| Hybrid regression | 7 | BM25+RRF breakage |

**Retrieval accuracy in CI: 100% (99/99).**

I'd rather fail a PR than fail a payroll question.

---

## Conformance: prove it, don't pitch it

From day one the bet was: **publish rules + tests anyone can run**.

```bash
pip install -r conformance/requirements.txt
python -m conformance spec          # offline OpenAPI contract
python -m conformance check --url http://localhost:8080
```

If your deployment is [Grounded-compatible](https://github.com/kantik001/grounded-llm/blob/main/docs/en/rfcs/RFC-0001-grounded-compatible.md), these pass without forking my repo.

v0.3 didn't replace the spec. It made the reference implementation **harder to dismiss as a demo**.

---

## Quick start (15 minutes, not 15 days)

```bash
git clone https://github.com/kantik001/grounded-llm.git
cd grounded-llm
cp .env.example .env

# Cloud LLM (default)
# LLM_API_KEY=...

# Or local CPU inference:
# LLM_PROVIDER=ollama
# docker compose --profile ollama up -d --build

docker compose up -d --build
python scripts/reindex_rag.py

pip install -r conformance/requirements.txt
python -m conformance spec
python -m conformance check --url http://localhost:8080
```

| Endpoint | URL |
|----------|-----|
| Web UI | http://localhost/ |
| API | http://localhost:8080 |
| Metrics | http://localhost:8080/metrics |
| gRPC Retriever | `localhost:50051` |

**Container images:**

```bash
docker pull ghcr.io/kantik001/grounded-llm-server:0.3.0
docker pull ghcr.io/kantik001/grounded-llm-python:0.3.0
docker pull ghcr.io/kantik001/grounded-llm-webapp:0.3.0
```

Template packs (HR, IT Support, Legal FAQ):

```bash
python scripts/init_pack.py install hr
python scripts/reindex_rag.py
```

---

## What I'd show in a demo (screenshot checklist)

If you're writing your own post or recording a Loom, this order hits hard:

1. **Architecture diagram** — the "aha" slide (use the PNG above)
2. **Chat with citations** — filename + chunk in the answer
3. **Out-of-scope refusal** — question not in KB → no hallucination
4. **`X-Cache: HIT`** on repeat question — infra credibility
5. **Conformance CLI green** — spec isn't vaporware
6. **CI retrieval gate** — 99/99 on GitHub Actions
7. **`LLM_PROVIDER=ollama`** — same API, local model

---

## Where this sits vs. the industry (still true in v0.3)

| Product | Focus | Grounded LLM difference |
|---------|-------|-------------------------|
| **NotebookLM** | Research / consumer grounding | Enterprise on-prem, API contract, CI gates |
| **Vertex AI Search** | Managed cloud retrieval | Self-hosted, MIT core, conformance badge |
| **LangChain / agent frameworks** | Composition flexibility | Narrow standard: **cited document Q&A** with verify + eval |

I'm not competing on "build any agent." I'm saying: when procurement asks *"is your internal assistant grounded and testable?"* — there should be a **published spec, a CLI, and a retrieval gate** — not a vendor slide.

---

## What's next

- **External conformance adopters** — run the CLI on your deploy, open issues for Spec v2
- **grounded-agent** — ReAct loop on top of gRPC Retriever + MCP Gateway
- **Load tests** — prove the cache + hybrid path under concurrency

If you work on **enterprise RAG, on-prem LLM, or OSS conformance** — I'd genuinely value feedback on [RFC-0001](https://github.com/kantik001/grounded-llm/blob/main/docs/en/rfcs/RFC-0001-grounded-compatible.md).

---

## Call to action

1. **Try v0.3.0:** [Release notes](https://github.com/kantik001/grounded-llm/releases/tag/v0.3.0) · `docker compose up` or pull GHCR `:0.3.0`
2. **Star / watch** if enterprise grounding interests you: [github.com/kantik001/grounded-llm](https://github.com/kantik001/grounded-llm)
3. **Run conformance** on your stack and tell me what belongs in Spec v2
4. **Contribute an eval case** when you fix a retrieval bug — see [GOOD_FIRST_ISSUES.md](https://github.com/kantik001/grounded-llm/blob/main/GOOD_FIRST_ISSUES.md)
5. **Read Part 1** if you missed the origin story: [Building an open standard for grounded document assistants](https://dev.to/kantik001/building-an-open-standard-for-grounded-document-assistants-2h6e)

---

## Summary

| Question | Answer |
|----------|--------|
| What is it? | Open platform + Spec v1 for cited, verified document assistants |
| What changed since Part 1? | Local LLMs, Redis caches, gRPC Retriever, 99-eval gate, GHCR images |
| What is it not? | Agent builder, ChatGPT clone, Glean competitor |
| Why the diagram? | Because enterprise RAG is a **system** — and systems deserve diagrams you want to show people |
| What's the bet? | Checkable standard any team can implement on their own infra |

I started with my father's horticulture papers.

v0.1 was the manifesto.

**v0.3 is the platform I'd deploy in a real pilot.**

---

*MIT · [Grounded Spec v1](https://github.com/kantik001/grounded-llm/blob/main/docs/en/spec/GROUNDED_SPEC_v1.md) · [Landing](https://kantik001.github.io/grounded-llm/) · [Architecture source PNG](https://github.com/kantik001/grounded-llm/blob/main/docs/assets/architecture.png)*
