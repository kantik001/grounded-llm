# Kantemir Satibalov

**AI Platform Engineer · RAG / LLM Systems**

Sochi, Russia · Open to remote (EU/US-friendly time zones)  
GitHub: [github.com/kantik001](https://github.com/kantik001) · DEV: [dev.to/kantik001](https://dev.to/kantik001)  
Email: [kantik001@yandex.ru]

---

## Summary

AI Platform / RAG engineer building **production document-grounded assistants** with measurable retrieval quality, source citations, and on-prem deploy. Author of **Grounded LLM v0.2.0** (MIT) — reference implementation of an open spec for cited, verified answers, with production fail-fast controls and CI security/load gates. Strong in hybrid retrieval (BM25 + dense + RRF), eval gates in CI, Go + Python service split, and enterprise-ready API/auth/deploy. Commercial experience: AI backend at **ADEPT** (IT solutions for construction). Technical writer on [dev.to](https://dev.to/kantik001) (hybrid RAG, eval-first quality).

---

## Experience

### Independent Project — Grounded LLM Platform

**AI Platform Engineer** · Feb 2025 – Present · Remote

Open-source platform to deploy **cited, verified document assistants** in days — not a generic chatbot wrapper.

- Shipped **Grounded LLM v0.2.0** ([release](https://github.com/kantik001/grounded-llm/releases/tag/v0.2.0)): production hardening on top of **Grounded Spec v1** — citations, numeric verify, conformance CLI, and **89-case retrieval eval gate** in GitHub Actions
- Hardened deploy path: `GROUNDED_ENV=production` fail-fast (secrets, no mocks/default DB password), Gunicorn for Python RAG, graceful shutdown, upload content sniffing, nginx API rate limits, `docker-compose.prod.yml`
- Added CI quality gates: **Trivy** image scan (CRITICAL/HIGH), concurrent **load smoke**, multi-tenant isolation regression tests (`ALLOWED_TENANTS` + per-tenant KB paths)
- Built **Go + Python** architecture: Go handles auth (Telegram / API key / OIDC admin), sessions, LLM orchestration, post-answer numeric verify, admin; Python serves retrieval-only `/rag/context`
- Implemented **hybrid retrieval**: BM25 sparse index + dense embeddings + **RRF** fusion; pluggable vector backends (**Chroma, Qdrant, pgvector**); optional keyword / cross-encoder rerank
- Delivered multi-tenant **REST API `/api/v1`**, OpenAPI, SSE streaming, Python SDK/CLI; domain template packs (HR, IT Support, Legal FAQ)
- Enterprise paths: RBAC, OIDC SSO, audit log, per-tenant quotas, async reindex; deploy via **Docker Compose, Helm/K8s, Terraform** (AWS/GCP/Azure); GHCR release images
- Ingest connectors: SharePoint, Google Drive, Confluence (live + export)
- Published platform story: [Building an open standard for grounded document assistants](https://dev.to/kantik001/building-an-open-standard-for-grounded-document-assistants-2h6e)

**Links:** [Repo](https://github.com/kantik001/grounded-llm) · [Landing](https://kantik001.github.io/grounded-llm/) · [Spec & benchmark docs](https://github.com/kantik001/grounded-llm/tree/main/docs/en)

---

### Gardener's Assistant — Vertical RAG Case Study

**Independent project** · 2024 – Present · Remote

Production-style **grounded RAG** for horticulture — separate codebase, same engineering patterns (not a bundled domain pack in Grounded LLM).

- Indexed ~**500 scientific articles** → ~**14.5k chunks**; hybrid **Chroma + BM25 + RRF + cross-encoder rerank**; query expansion via domain glossary
- **68/68** retrieval regression cases; eval-first workflow before LLM integration
- Shipped Telegram Mini App + browser client; Go orchestration + Python RAG + PostgreSQL sessions
- Case study & demos: [grounded_horticulture_en](https://github.com/kantik001/grounded_horticulture_en) · [DEV series](https://dev.to/kantik001)

---

### ADEPT

**AI Backend Engineer** · June 2024 – February 2025 · Remote

IT integrator delivering **digital solutions for the construction industry** (operational workflows, internal tools, data pipelines).

- Built **REST APIs in Python (FastAPI)** integrating AI-based classification into construction-domain workflows
- Reduced manual processing of incoming requests
- Worked with **PostgreSQL** (schema design, migrations, query tuning for production workloads)
- **CI/CD (GitLab) + Docker**
- Collaborated with frontend and product teams; code review in a production codebase

---

### Gallery Maykop

**Manager / Curator** · `[2018]` – `[2024]` · Sochi, Russia

- Managed exhibition catalogs, artifact descriptions, and document workflows
- Structured large volumes of domain text — foundation for later document-grounded RAG work

---

## Selected Technical Writing

| Topic | Link |
|-------|------|
| Open spec for grounded assistants | [dev.to](https://dev.to/kantik001/building-an-open-standard-for-grounded-document-assistants-2h6e) |
| Hybrid search (BM25 + RRF) | [dev.to](https://dev.to/kantik001/vector-search-kept-missing-rootstock-codes-so-i-went-hybrid-37li) |
| Eval-first RAG (68 cases) | [dev.to](https://dev.to/kantik001/68-questions-before-a-single-token-eval-first-rag-1g6g) |
| Passion Edition — horticulture RAG | [dev.to](https://dev.to/kantik001/my-father-wrote-the-papers-i-built-a-rag-assistant-so-growers-can-query-them-safely-1hi) |

---

## Skills

**Languages & runtimes:** Python, Go, SQL  
**RAG / LLM:** Hybrid retrieval (BM25, dense embeddings, RRF, reranking), chunking, citations, adversarial eval, conformance testing  
**Data & search:** PostgreSQL, pgvector, Chroma, Qdrant, Redis  
**API & platform:** REST, OpenAPI, SSE, multi-tenant auth (API key, OIDC), RBAC, audit  
**Frameworks:** FastAPI, Flask/Gunicorn, Gin  
**DevOps:** Docker, Docker Compose, Kubernetes, Helm, Terraform, GitHub Actions, GitLab CI, Trivy  
**Observability:** Prometheus (basic metrics)  
**Other:** Linux, Nginx, Git

---

## Education

**RUDN University** (Peoples' Friendship University of Russia) · Moscow

- **Master's degree, Law** · [2014-2016]  
- **Bachelor's degree, Economics** · [2010-2014]

---

## Languages

- **Russian** — Native  
- **English** — B2
