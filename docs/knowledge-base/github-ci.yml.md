# Разбор: `.github/workflows/ci.yml` / CI workflow

**Исходный файл / Source:** `.github/workflows/ci.yml`

---

## Когда запускается / Triggers

```yaml
on:
  push:
    branches: [master, main, "feature/**"]
  pull_request:
    branches: [master, main]
```

Результат: вкладка **Actions** на GitHub.

---

## Три job параллельно / Three parallel jobs

| Job | Проверяет |
|-----|-----------|
| `go-test` | `go test ./...` в `server/` |
| `python-test` | `pytest tests/` |
| `docker-build` | сборка `Dockerfile.server`, `Dockerfile.webapp`, `Dockerfile.python` |

Если падает хотя бы один — workflow failed.

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

Зависимости тестов включают `langchain-community`, `pypdf`, `docx2txt` (для `rag/document_loaders.py`).

**Не тестируется здесь:** Chroma end-to-end, LLM, полный Docker compose.

---

## Job `docker-build`

```bash
docker build -f Dockerfile.server -t grounded-llm-server:ci .
docker build -f Dockerfile.webapp -t grounded-llm-webapp:ci .
docker build -f Dockerfile.python -t grounded-llm-python:ci .
```

Образы не пушатся в registry — только проверка сборки.

---

## Локально повторить CI

```powershell
cd server; go mod tidy; go test ./...
pip install -r tests/requirements-test.txt
pytest tests/ -v
docker build -f Dockerfile.server -t test-server .
docker build -f Dockerfile.python -t test-python .
```

---

## Чего пока нет в CI

- `docker compose up` / smoke E2E
- eval RAG end-to-end с LLM
- деплой (CD)

---

## Частые причины падений

| Причина | Решение |
|---------|---------|
| `ModuleNotFoundError: langchain_community` | обновить `tests/requirements-test.txt` |
| нет `config/domains.json` в git | файл должен быть в репозитории |
| не закоммичен `go.sum` | `go mod tidy` и commit |
