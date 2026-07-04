# Contributing to Grounded LLM

Thank you for your interest in contributing. Grounded LLM is an open platform for **document-grounded assistants** — citations, verification, and measurable retrieval quality.

## Before you start

1. Read [PLATFORM_VISION.md](PLATFORM_VISION.md) to understand scope and non-goals.
2. Check [open issues](https://github.com/kantik001/grounded-llm/issues) and [docs/en/ROADMAP.md](docs/en/ROADMAP.md) for planned work.
3. For security issues, see [SECURITY.md](SECURITY.md) — **do not open public issues for vulnerabilities**.

## Development setup

```bash
git clone https://github.com/kantik001/grounded-llm.git
cd grounded-llm
cp .env.example .env
# Set LLM_API_KEY (OpenAI-compatible). For local browser dev: TELEGRAM_AUTH_DISABLED=true

docker compose up -d --build
python scripts/reindex_rag.py
```

| Service | URL |
|---------|-----|
| Web App | http://localhost/ |
| Go API | http://localhost:8080/health |
| Python RAG | http://localhost:5000/health |

See [docs/en/DEPLOY.md](docs/en/DEPLOY.md) and [docs/en/knowledge-base/](docs/en/knowledge-base/) for module-level details.

## Running tests

```bash
make test                 # Go + Python unit tests
make eval-retrieval-ci    # Full retrieval gate (reindex + eval, same as CI)
make smoke                # Smoke API against localhost:8080
pip install -e "sdk/python[dev]"   # Python SDK + CLI
grounded-llm chat "How many vacation days?"   # after server is up
```

**CI runs on every push/PR:**

| Job | Scope |
|-----|-------|
| `go-lint` | golangci-lint on `server/` |
| `go-test` | Go unit tests + coverage |
| `python-lint` | Ruff on `rag/`, `api/`, `tests/`, `scripts/` |
| `python-test` | pytest + coverage |
| `openapi-validate` | OpenAPI 3.0 spec validation |
| `eval-baseline-validate` | JSONL structure validation |
| `eval-retrieval-gate` | Reindex + retrieval eval (all suites) |
| `smoke-api` | Health, domains, session, **full `/message` path** (with `LLM_MOCK` + `RAG_MOCK`) |
| `docker-build` | Build all Docker images |

Dependabot opens weekly dependency PRs.

> **CodeQL:** optional manual/weekly workflow (`.github/workflows/codeql.yml`). Not gated on PRs. Turn on [Code scanning](https://docs.github.com/en/code-security/code-scanning) in repo settings, then set `upload: true` and add `pull_request` trigger if desired.

**Recommended local checks before a PR:**

```bash
cd server && golangci-lint run ./...
ruff check rag api tests scripts
make test
make eval-retrieval-ci   # if you changed retrieval, config, eval, or KB data
bash scripts/smoke.sh http://127.0.0.1:8080   # with server running; set LLM_MOCK=true RAG_MOCK=true
```

Changes to `rag/`, `config/`, `eval/`, or `data/` that affect retrieval **must pass** `make eval-retrieval-ci` locally before opening a PR.

## Pull request process

1. Fork the repository and create a branch from `main` (e.g. `feature/my-change` or `fix/issue-123`).
2. Make focused changes — one logical change per PR when possible.
3. Add or update tests for behavior changes.
4. Update documentation if you change APIs, config, or deploy steps.
5. Fill out the PR template checklist.
6. Ensure CI is green before requesting review.

We review PRs as time allows. Be patient — this is a community-driven project.

## Code guidelines

### Go (`server/`)

- Run tests: `cd server && go test ./...`
- Match existing patterns: Gin handlers, middleware, env-based config.
- Keep API contracts stable or document breaking changes in CHANGELOG.

### Python (`rag/`, `api/`, `tests/`)

- Run tests: `pytest tests/ -v`
- Follow existing module layout; avoid large new dependencies without discussion.

### Template packs (`packs/`, `config/`, `eval/`)

- New domain packs should include: config, sample data, eval baseline JSONL.
- See [packs/README.md](packs/README.md) and [domain-pack-template/](domain-pack-template/).

### Documentation

- Primary language: **English** in root files and `docs/en/`.
- Russian docs in `docs/ru/` are legacy locale mirrors — update EN first.

## Commit messages

Use clear, descriptive messages:

```
fix(rag): use $and filter for multi-tenant Chroma queries
feat(packs): add legal FAQ template scaffold
docs: update API examples for streaming
test: add verifier edge cases for decimal numbers
```

## What to contribute

**Great first contributions:**

- Documentation fixes and clarifications
- Eval baseline cases (`eval/*.jsonl`)
- Locale bundles (`config/locales/`)
- Template packs (HR, IT, legal FAQ)
- Test coverage for edge cases

**Needs discussion first:**

- New core dependencies
- Breaking API changes
- Scope expansion beyond document-grounded assistants

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you agree to uphold it.

## Questions

- **Bug reports and features:** [GitHub Issues](https://github.com/kantik001/grounded-llm/issues)
- **Architecture and design:** See [HIRING.md](HIRING.md) and [docs/en/ARCHITECTURE.md](docs/en/ARCHITECTURE.md)
- **Security:** [SECURITY.md](SECURITY.md)

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
