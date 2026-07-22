# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Helm: configurable probes + Python `startupProbe` (model/index warm-up); timeouts/`failureThreshold` for server, python, postgres, webapp
- `scripts/backup_postgres_smoke.sh` + `make backup-smoke` + CI job (pg_dump → restore round-trip)
- Trivy scan for Python RAG image in CI (CRITICAL; torch stack often has unfixed HIGH)
- HR paraphrase + adversarial near-miss retrieval cases (eval total **99**)
- Local Compose default `RAG_RERANKER=keyword` for stronger demo ranking (override via env; CI stays `none` unless set)

## [0.2.0] - 2026-07-22

Production hardening and CI quality gates for safer deploys.

### Added

- **Production hardening:** `GROUNDED_ENV=production` fail-fast checks (admin password, `RAG_SERVICE_TOKEN`, `ADMIN_SECRET`, no default DB password, no mock modes)
- **`docker-compose.prod.yml`:** required secrets, no public Python/Postgres ports, resource limits
- **Gunicorn** for Python RAG (`Dockerfile.python`) instead of Flask development server
- Graceful shutdown on SIGINT/SIGTERM for Go server
- Knowledge upload content sniffing (PDF/DOCX magic, UTF-8 TXT)
- nginx `limit_req` on `/api/`
- Multi-tenant isolation tests (`ALLOWED_TENANTS` + per-tenant KB paths)
- Trivy image scan (CRITICAL/HIGH) on server and webapp CI images
- Concurrent load smoke (`scripts/load_smoke.sh`, optional `scripts/load_smoke.js` for k6)

### Changed

- Local Compose binds Python RAG to `127.0.0.1:5000` only (not all interfaces)
- Python CORS disabled by default (internal service); optional `PYTHON_CORS_ORIGINS`
- SECURITY.md / NETWORK_SECURITY.md aligned with token-based Go ↔ Python auth
- Go toolchain / golangci-lint aligned for Go 1.25 (pgx v5.10)

## [0.1.0] - 2026-07-14

First tagged open-source release — reference implementation of the **Grounded** standard for document-grounded assistants (citations, numeric verify, measurable retrieval quality, on-prem deploy).

### Platform core

- Go orchestration (auth, sessions, LLM, verify, admin) + Python retrieval service
- Multi-tenant REST API `/api/v1`, OpenAPI, SSE streaming, Python SDK/CLI
- Citations in chat, numeric verify layer, admin upload/reindex, feedback analytics
- Enterprise: RBAC, OIDC SSO, audit log, per-tenant quotas, async reindex jobs
- Optional SaaS path: signup API, Stripe webhook/checkout, plan tiers

### Standard & quality (Horizon 1)

- [Grounded Spec v1](docs/en/spec/GROUNDED_SPEC_v1.md) — normative API contract
- Conformance CLI: `python -m conformance` (spec, http, retrieval, check, all)
- [RFC-0001 Grounded-compatible](docs/en/rfcs/RFC-0001-grounded-compatible.md)
- [STANDARD_STRATEGY.md](docs/en/STANDARD_STRATEGY.md), [ECOSYSTEM.md](docs/en/ECOSYSTEM.md)
- Eval harness: 89+ retrieval cases (HR, IT, Legal, adversarial, hybrid) with CI gate
- Adversarial E2E smoke against `/message`

### Retrieval

- Vector backends: **Chroma** (default), **Qdrant**, **pgvector** (`VECTOR_STORE`)
- **Hybrid retrieval:** BM25 sparse + dense embeddings + RRF (`RAG_RETRIEVAL_MODE=hybrid`)
- Optional rerank: keyword overlap or cross-encoder (`RAG_RERANKER`)
- Stable `chunk_id` metadata for dense/sparse fusion

### Templates & connectors

- Domain packs: HR, IT Support, Legal FAQ (`packs/registry.yaml`)
- Ingest connectors: SharePoint, Google Drive, Confluence (live + export)
- Embeddable widget (`webapp/embed.html`)

### Deploy

- Docker Compose, Helm chart (`deploy/helm/grounded-llm/`)
- Terraform references: AWS, Azure, GCP
- Postgres `pgvector/pgvector:pg16` image in reference deploy
- GHCR release workflow on version tags

### Docs

- Architecture, deploy, security brief, trust center, benchmark guide
- English-first docs (`docs/en/`) + Russian hub (`docs/ru/`)
- GitHub Pages landing (`site/`)

[Unreleased]: https://github.com/kantik001/grounded-llm/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/kantik001/grounded-llm/releases/tag/v0.2.0
[0.1.0]: https://github.com/kantik001/grounded-llm/releases/tag/v0.1.0
