# `.github/workflows/ci.yml` — CI

**Источник:** `.github/workflows/ci.yml` · карта: [`.github/AUTOMATION.md`](../../../.github/AUTOMATION.md)  
**EN:** [github-ci.yml.md](../../en/knowledge-base/github-ci.yml.md)

---

## Триггеры

- **push** на `main`, `master`, `feature/**`, `refactor/**`, `chore/**`, `fix/**`, `docs/**`
- **pull_request** на `main` / `master`
- **workflow_dispatch**

Env: Go **1.25**, Python **3.11**, `DOMAINS_CONFIG_PATH`, `LOCALES_ROOT`.

---

## Jobs

| Job | Что проверяет |
|-----|----------------|
| `go-lint` | golangci-lint в `server/` |
| `go-test` | `go test ./...` + coverage (`LLM_MOCK`, `RAG_MOCK`) |
| `python-lint` | Ruff + sync gRPC stubs |
| `proto-lint` | `buf lint` + breaking vs `main` на PR |
| `conformance-spec` | offline OpenAPI (`python -m conformance spec`) |
| `python-test` | pytest + **70%** coverage на `rag/` |
| `sdk-test` | `sdk/python` |
| `eval-baseline-validate` | схема JSONL |
| `eval-retrieval-gate` | reindex + **99** кейсов (hybrid) |
| `smoke-api` | Postgres + mock server + adversarial E2E + HTTP conformance |
| `docker-build` | 3 образа + Trivy |
| `helm-lint` | Helm chart |
| `secret-scan` | gitleaks |

Отдельно: `eval-llm-nightly`, `codeql`, `release`.

---

## Retrieval gate

1. Prefetch `intfloat/multilingual-e5-small`  
2. `python scripts/reindex_rag.py` (`RAG_RETRIEVAL_MODE=hybrid`)  
3. `bash scripts/ci_eval_retrieval.sh` — все сьюиты из [eval/README.md](../../../eval/README.md)

Это бенчмарк из README и [BENCHMARK.md](../BENCHMARK.md).

---

## Локальный паритет

```bash
make test
make eval-retrieval-ci
python -m conformance spec
cd server && golangci-lint run ./...
```
