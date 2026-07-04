# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- OSS governance: `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `SECURITY.md`
- GitHub issue and pull request templates
- This changelog

### Changed

- README: community-focused layout; maintainer section condensed
- Root files: English-first comments in `Makefile` and `.env.example`
- `PROJECT_STRUCTURE.md`: updated paths and documentation links

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
