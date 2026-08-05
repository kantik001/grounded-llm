# `.github/` — automation for this repository

> Not the repository homepage README. GitHub prefers `.github/README.md` over the root `README.md`, so this map lives here instead.

| Path | Role |
|------|------|
| `workflows/ci.yml` | PR/push quality gate (lint, tests, retrieval eval, images, smoke) |
| `workflows/secret-scan.yml` | gitleaks |
| `workflows/codeql.yml` | CodeQL SAST |
| `workflows/release.yml` | Release / GHCR |
| `workflows/pages.yml` | Docs site |
| `workflows/eval-llm-nightly.yml` | Nightly eval with real LLM |
| `actions/` | Shared composite actions used by CI |
| `CODEOWNERS` | Default review owners |
| `dependabot.yml` | Grouped dependency updates |
| `ISSUE_TEMPLATE/` | Bug / feature / eval regression |
| `pull_request_template.md` | PR checklist |

## CI layout (`workflows/ci.yml`)

Jobs stay in one workflow named **CI** so branch-protection required checks keep stable names.

Shared pieces:
- top-level `env` for domains/locales/HF cache paths
- `.github/actions/setup-python-ci`
- `.github/actions/setup-go-ci`
- `.github/actions/cache-hf-e5`
- `scripts/ci_start_mock_server.sh` for mock Go server in smoke/load jobs

Push CI runs on `main`/`master` and common prefix branches: `feature/**`, `refactor/**`, `chore/**`, `fix/**`, `docs/**`.
Pull requests targeting `main`/`master` always run the full gate.
