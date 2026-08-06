# `.github/workflows/ci.yml` — CI workflow

**Source:** `.github/workflows/ci.yml` · map: [`.github/AUTOMATION.md`](../../../.github/AUTOMATION.md)

---

## Triggers

- **push** to `main`, `master`, `feature/**`, `refactor/**`, `chore/**`, `fix/**`, `docs/**`
- **pull_request** to `main` / `master`
- **workflow_dispatch**

Shared env: Go **1.25**, Python **3.11**, `DOMAINS_CONFIG_PATH`, `LOCALES_ROOT`.

---

## Jobs (summary)

| Job | What it checks |
|-----|----------------|
| `go-lint` | golangci-lint on `server/` |
| `go-test` | `go test ./...` with coverage (`LLM_MOCK`, `RAG_MOCK`) |
| `python-lint` | Ruff + gRPC stub sync (`gen_retriever_grpc.py --check`) |
| `proto-lint` | `buf lint` + breaking vs `main` on PRs |
| `conformance-spec` | offline OpenAPI spec (`python -m conformance spec`) |
| `python-test` | pytest + **70%** coverage on `rag/` |
| `openapi-validate` | alias → covered by `conformance-spec` |
| `sdk-test` | `sdk/python` pytest + Ruff |
| `eval-baseline-validate` | JSONL schema for all eval suites |
| `eval-retrieval-gate` | reindex + **99** retrieval cases (hybrid + keyword rerank) |
| `smoke-api` | Postgres + mock server: health, `/message`, adversarial E2E, HTTP conformance |
| `docker-build` | build 3 images + Trivy (server/webapp HIGH; python CRITICAL) |
| `helm-lint` | `helm lint` on chart |
| `secret-scan` | gitleaks |

Nightly / separate workflows: `eval-llm-nightly`, `codeql`, `release` — see `.github/workflows/`.

---

## Retrieval gate (product quality)

`eval-retrieval-gate`:

1. Prefetch `intfloat/multilingual-e5-small`
2. `python scripts/reindex_rag.py` (`RAG_RETRIEVAL_MODE=hybrid`)
3. `bash scripts/ci_eval_retrieval.sh` — all suites in [eval/README.md](../../eval/README.md) (**99** cases)

This is the benchmark cited in README and [BENCHMARK.md](../BENCHMARK.md).

---

## What CI does **not** cover

- Live LLM provider billing (use `LLM_MOCK` / nightly)
- Full Docker Compose stack on every push (smoke uses mock server + Postgres)
- Chroma persistence across jobs (ephemeral reindex each run)

---

## Local parity

```bash
make test
make eval-retrieval-ci
python -m conformance spec
cd server && golangci-lint run ./...
```
