# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Phase 1 engineering bar:** golangci-lint, Ruff, CodeQL, Dependabot, OpenAPI validation in CI
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

- README: SDK quickstart and product evidence links

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

[Unreleased]: https://github.com/kantik001/grounded-llm/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/kantik001/grounded-llm/releases/tag/v0.1.0
