# Comparison — Grounded LLM vs alternatives

Honest positioning for architects and product reviewers. We optimize for **trust, deployability, and measured RAG quality** — not maximum feature count.

---

## At a glance

| | Grounded LLM | Tutorial RAG repo | LangChain / LlamaIndex | Dify / Flowise | Glean (SaaS) |
|---|:---:|:---:|:---:|:---:|:---:|
| **Primary goal** | Cited internal assistants | Learning | Library | Visual AI apps | Enterprise search |
| **Deploy on-prem** | ✅ First-class | Varies | DIY | Partial | ❌ |
| **Citations in product** | ✅ | Sometimes | DIY | Sometimes | ✅ |
| **Numeric verify layer** | ✅ | Rare | DIY | Rare | Unknown |
| **Retrieval eval in CI** | ✅ | Rare | Rare | Rare | N/A |
| **Multi-tenant API** | ✅ | Rare | DIY | ✅ | ✅ |
| **Template packs (HR, IT)** | ✅ | ❌ | ❌ | Workflows | Vertical |
| **Open source core** | MIT | Varies | MIT | Apache | ❌ |
| **Agent / workflow builder** | ❌ By design | ❌ | ✅ | ✅ | Partial |

---

## When to choose Grounded LLM

- You need **policy/handbook Q&A** with citations inside **your** infrastructure  
- IT requires a **security brief**, audit log, and on-prem Docker/K8s path  
- You want **regression tests on retrieval** before merge, not only manual chat tests  
- You ship **repeatable assistants** via template packs (HR, IT support), not one-off demos  

## When to choose something else

| Need | Better fit |
|------|------------|
| Arbitrary agent graphs, tool use, code execution | LangChain, LangGraph, Dify |
| Maximum OSS component ecosystem | LlamaIndex + your own glue |
| Turnkey cloud search across SaaS apps | Glean, Microsoft Copilot |
| Fastest “hello world” notebook | Tutorial RAG (smaller scope) |
| Visual non-developer builder | Dify, Flowise |

---

## Differentiators (deep)

### 1. Eval-driven retrieval

JSONL suites (`eval/*.jsonl`) run in CI (`eval-retrieval-gate`) without LLM token cost. Separates retrieval bugs from generation bugs.

### 2. Verify layer

Post-LLM check: numeric claims in answers must appear in retrieved chunks. Imperfect but **testable** — important for HR/policy use cases.

### 3. Split Go / Python architecture

Go: auth, sessions, LLM, verify, admin API. Python: embeddings + Chroma only. Clear trust boundary and scaling story.

### 4. Platform vs pack

Core repo stays stable; customers add `config/` + `data/` + eval per use case. See [packs/README.md](../../packs/README.md).

---

## What we explicitly do not compete on

- Foundation model training or fine-tuning  
- Universal chat without a knowledge base  
- Consumer mobile apps  
- Feature parity with vertical SaaS search  

---

See also: [PLATFORM_VISION.md](../../PLATFORM_VISION.md) · [ROADMAP.md](./ROADMAP.md)
