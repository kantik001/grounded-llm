# `.github/workflows/ci.yml` — CI workflow

**Source:** `.github/workflows/ci.yml`

---

## Triggers

```yaml
on:
  push:
    branches: [master, main, "feature/**"]
  pull_request:
    branches: [master, main]
```

Results appear under **Actions** on GitHub.

---

## Three parallel jobs

| Job | Checks |
|-----|--------|
| `go-test` | `go test ./...` in `server/` |
| `python-test` | `pytest tests/` |
| `docker-build` | build `Dockerfile.server`, `Dockerfile.webapp`, `Dockerfile.python` |

If any job fails — workflow failed.

---

## Job `go-test`

- Go **1.23**
- `go mod tidy` + `go test -v -count=1 ./...`
- Env: `DOMAINS_CONFIG_PATH=${{ github.workspace }}/config/domains.json`

---

## Job `python-test`

- Python **3.11**
- `pip install -r tests/requirements-test.txt`
- `pytest tests/ -v --tb=short`
- Env: `DOMAINS_CONFIG_PATH=config/domains.json`

Test deps include `langchain-community`, `pypdf`, `docx2txt` (for `rag/document_loaders.py`).

**Not tested here:** Chroma end-to-end, LLM, full Docker compose.

---

## Job `docker-build`

```bash
docker build -f Dockerfile.server -t grounded-llm-server:ci .
docker build -f Dockerfile.webapp -t grounded-llm-webapp:ci .
docker build -f Dockerfile.python -t grounded-llm-python:ci .
```

Images are not pushed — build verification only.

---

## Reproduce CI locally

```powershell
cd server; go mod tidy; go test ./...
pip install -r tests/requirements-test.txt
pytest tests/ -v
docker build -f Dockerfile.server -t test-server .
docker build -f Dockerfile.python -t test-python .
```

---

## Not in CI yet

- `docker compose up` / smoke E2E
- eval RAG end-to-end with LLM
- deploy (CD)

---

## Common failure causes

| Cause | Fix |
|-------|-----|
| `ModuleNotFoundError: langchain_community` | update `tests/requirements-test.txt` |
| missing `config/domains.json` | file must be in repo |
| unstaged `go.sum` | `go mod tidy` and commit |
