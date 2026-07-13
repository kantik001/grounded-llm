# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Phase 1 engineering bar:** golangci-lint, Ruff, Dependabot, OpenAPI validation in CI
- **Mock modes for CI:** `LLM_MOCK` and `RAG_MOCK` for deterministic smoke/E2E without external APIs
- **Release workflow:** GitHub Release + GHCR images on `v*.*.*` tags
- **Expanded OpenAPI:** public endpoints (health, metrics, domains, branding, onboarding) + chat schemas
- Go coverage reporting and Python `pytest-cov` in CI
- Full `/message` smoke test path (session → cited answer with verify)

### Changed

- Smoke script covers metrics, branding, and message flow
- CI jobs: `go-lint`, `python-lint`, `openapi-validate`

### Added (Phase 2 — adoption)

- **Python SDK + CLI:** `sdk/python/` (`pip install -e "sdk/python"`, command `grounded-llm`)
- Product docs: case study, comparison, analytics guide, SDK quickstart, demo video script
- Blog: [retrieval eval gate in CI](docs/en/blog/retrieval-eval-gate-in-ci.md)
- [GOOD_FIRST_ISSUES.md](GOOD_FIRST_ISSUES.md), [examples/python/chat_basic.py](examples/python/chat_basic.py)
- Nightly LLM E2E workflow + `scripts/llm_e2e_smoke.sh`
- CI job `sdk-test`

### Changed (Phase 2)

- README: expanded architecture diagram, SDK quickstart and product evidence links
- CodeQL moved off PR checks (manual/weekly only; enable upload when Code scanning is on)

### Added (Phase 3 — enterprise scale)

- **Readiness probes:** `GET /ready` on Go (DB + Python RAG) and Python (`/ready` with chroma/data checks)
- **`RAG_SERVICE_TOKEN`:** internal auth header `X-RAG-Service-Token` for Go → Python calls
- **Retention worker:** `MESSAGE_RETENTION_DAYS`, `SESSION_RETENTION_DAYS`, `RETENTION_INTERVAL_HOURS`
- **Helm chart:** `deploy/helm/grounded-llm/` with probes, secrets, PVCs
- **Enterprise docs:** [TRUST_CENTER.md](docs/en/TRUST_CENTER.md), [BACKUP_RESTORE.md](docs/en/BACKUP_RESTORE.md), [K8S_DEPLOY.md](docs/en/K8S_DEPLOY.md), [NETWORK_SECURITY.md](docs/en/NETWORK_SECURITY.md)
- **nginx CSP** headers on webapp shell

### Changed (Phase 3)

- Docker Compose server healthcheck uses `/ready` instead of `/health`
- Smoke script checks `/ready`

### Planned (Phase 4 — spec & trust, prep on `docs/phase-4-prep`)

- [PHASE_4.md](docs/en/PHASE_4.md) implementation plan
- [API_DEPRECATION_POLICY.md](docs/en/API_DEPRECATION_POLICY.md)
- [COMPATIBILITY.md](docs/en/COMPATIBILITY.md)
- Conformance suite scaffold: [conformance/](conformance/)
- Adversarial eval pack: `eval/rag_adversarial_baseline.jsonl` (25 cases)
- Secret scanning workflow: `.github/workflows/secret-scan.yml`
- Tenant purge spec: [TENANT_PURGE.md](docs/en/TENANT_PURGE.md)

### Added (Phase 4 — spec & trust)

- **`expect_not_contains`** in RAG eval runner + adversarial retrieval gate
- **Tenant purge:** `DELETE /api/admin/tenants/:tenant_id?confirm=true` (admin RBAC, audit)
- **Adversarial E2E:** `eval/rag_adversarial_e2e.jsonl` + `scripts/run_adversarial_e2e.py` in CI smoke
- **Conformance HTTP** tests against running server (smoke job)
- **Smarter RAG mock** for out-of-scope adversarial cases in CI
- Unit tests: `tests/test_rag_eval_check.py`, `admin_tenant_purge_test.go`

### Added (Phase 5 — standard publication)

- [GROUNDED_SPEC_v1.md](docs/en/spec/GROUNDED_SPEC_v1.md) — normative Spec v1
- Conformance CLI: `python -m conformance` (`spec`, `http`, `retrieval`, `check`, `all`)
- [RFC.md](docs/en/RFC.md) + [RFC-0001 Grounded-compatible](docs/en/rfcs/RFC-0001-grounded-compatible.md)
- [STANDARD_STRATEGY.md](docs/en/STANDARD_STRATEGY.md) — five pillars, three horizons
- [BENCHMARK.md](docs/en/BENCHMARK.md) + `scripts/bench_report.py`
- [PHASE_5.md](docs/en/PHASE_5.md) — phase plan mapped to pillars

### Added (Phase 5b — site & release prep)

- GitHub Pages landing: [site/](../../site/) + `Deploy site` workflow
- Conformance CLI `--json` output for integrator pipelines
- [RELEASE.md](docs/en/RELEASE.md) — v0.3.0 tag checklist
- Release workflow includes conformance verification steps

### Added (Phase 6 — ecosystem scale)

- **Legal FAQ template pack:** `packs/legal_faq/` + [LEGAL_FAQ.md](docs/en/domain-packs/LEGAL_FAQ.md)
- **Vector store adapter:** `rag/vector_backend/` — Chroma default, Qdrant optional (`api/requirements-qdrant.txt`)
- **Hybrid retrieval:** `RAG_RETRIEVAL_MODE=hybrid` keyword rerank via `rag/hybrid_rank.py`
- **AWS Terraform reference:** `deploy/terraform/aws/reference/` + [TERRAFORM.md](docs/en/TERRAFORM.md)
- [VECTOR_STORE.md](docs/en/VECTOR_STORE.md), [PHASE_6.md](docs/en/PHASE_6.md)

### Added (Phase 7 — platform ecosystem)

- **Pack registry:** `packs/registry.yaml` + `init_pack.py registry --validate`
- **Cross-encoder rerank:** `rag/rerank.py`, `RAG_RERANKER=cross_encoder`
- **Ingest connectors:** `connectors/` + [CONNECTORS.md](docs/en/CONNECTORS.md), `scripts/sync_connector.py`
- **GCP Terraform reference:** `deploy/terraform/gcp/reference/`
- **Governance docs:** [GOVERNANCE.md](docs/en/GOVERNANCE.md), [PARTNER_CERTIFICATION.md](docs/en/PARTNER_CERTIFICATION.md)
- [PHASE_7.md](docs/en/PHASE_7.md)

### Added (Phase 8 — connectors & multi-cloud)

- **Export connectors:** `sharepoint_export`, `google_drive_export`, `confluence_export`
- **SharePoint Graph connector:** `connectors/sharepoint.py` (Microsoft Graph app-only)
- **Azure Terraform reference:** `deploy/terraform/azure/reference/`
- **Embeddable widget:** `webapp/embed.html` + [EMBED.md](docs/en/EMBED.md)
- **Site pack index:** `site/packs.json`, `scripts/build_site_data.py`
- **Pages workflow:** manual `workflow_dispatch` only (private repo on GitHub Free)
- [PHASE_8.md](docs/en/PHASE_8.md)

### Added (Phase 9 — launch & live connectors)

- **Google Drive API connector:** `connectors/google_drive.py`
- **Confluence REST connector:** `connectors/confluence.py`
- **Connector optional deps:** `api/requirements-connectors.txt`
- **Billing scaffold:** `config/plans.yaml`, [BILLING.md](docs/en/BILLING.md), [SAAS.md](docs/en/SAAS.md)
- **Launch playbook:** [LAUNCH.md](docs/en/LAUNCH.md)
- **Site packs section:** dynamic load from `packs.json`
- [PHASE_9.md](docs/en/PHASE_9.md)

### Added (Phase 10 — SaaS billing & signup)

- **Signup API:** `POST /api/v1/signup`, `GET /api/v1/plans`
- **Stripe webhook:** `POST /api/v1/billing/stripe/webhook` → tenant quotas
- **Tenant registry:** `config/tenants.json.example`, `server/tenant_registry.go`
- **Signup UI:** `webapp/signup.html`
- [PHASE_10.md](docs/en/PHASE_10.md)

### Added (Phase 11 — checkout & admin provisioning)

- **Stripe Checkout:** `POST /api/v1/billing/stripe/checkout`
- **Admin auto-provision:** signup creates `{tenant}-admin` in `ADMIN_USERS_FILE`
- **Paid plan flow:** starter quotas until Stripe webhook upgrades plan
- **Plans:** `stripe_price_id` in `config/plans.yaml`
- **Plans:** `stripe_price_id` in `config/plans.yaml`
- [PHASE_11.md](docs/en/PHASE_11.md)

### Changed (docs refresh)

- Updated ROADMAP, README, BILLING/SAAS, RELEASE, API examples for Phases 10–11 complete
- Russian ROADMAP points to English roadmap for phases 4–11

## [0.3.0] - TBD

Standard publication release: Spec v1, conformance CLI, adversarial eval, tenant purge, Helm, SDK adoption docs.

See `[Unreleased]` section above for full list (Phases 2–5b).

## [0.1.0] - 2026-07-05

Initial open-source release baseline.

### Added

- **Platform core:** Go orchestration (auth, sessions, LLM, verify, admin) + Python RAG (Chroma, embeddings)
- **Multi-tenant API:** `X-Tenant-ID`, `X-API-Key`, OpenAPI at `/api/v1/openapi.json`
- **SSE streaming:** `POST /message?stream=1`
- **Verify layer:** numeric claim verification against retrieved context
- **Template packs:** HR and IT Support (`packs/`, `scripts/init_pack.py`)
- **Eval harness:** JSONL baselines + retrieval gate in CI
- **Enterprise features:** RBAC, OIDC SSO, audit log, per-tenant quotas, async reindex, analytics dashboard
- **Reference UI:** Telegram Web App + admin (`webapp/`)
- **Documentation:** architecture, deploy, security brief, knowledge base (`docs/en/`, `docs/ru/`)

[Unreleased]: https://github.com/kantik001/grounded-llm/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/kantik001/grounded-llm/compare/v0.1.0...v0.3.0
[0.1.0]: https://github.com/kantik001/grounded-llm/releases/tag/v0.1.0
