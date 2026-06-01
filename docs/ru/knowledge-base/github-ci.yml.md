# `.github/workflows/ci.yml` — CI

**Исходник:** `.github/workflows/ci.yml`

---

## Когда запускается

```yaml
on:
  push:
    branches: [master, main, "feature/**"]
  pull_request:
    branches: [master, main]
```

Результат — вкладка **Actions** на GitHub.

---

## Три параллельных job

| Job | Проверяет |
|-----|-----------|
| `go-test` | `go test ./...` в `server/` |
| `python-test` | `pytest tests/` |
| `docker-build` | сборка трёх Dockerfile |

Падение любого job → workflow failed.

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

**Не проверяется:** Chroma end-to-end, LLM, полный compose.

---

## Job `docker-build`

Сборка образов server, webapp, python без push в registry.

---

## Локально как в CI

```powershell
cd server; go mod tidy; go test ./...
pip install -r tests/requirements-test.txt
pytest tests/ -v
docker build -f Dockerfile.server -t test-server .
```

---

## Чего пока нет в CI

- `docker compose up` / smoke E2E
- eval RAG с LLM
- деплой (CD)

---

## Частые причины падений

| Причина | Решение |
|---------|---------|
| `ModuleNotFoundError: langchain_community` | обновить `tests/requirements-test.txt` |
| нет `config/domains.json` | файл должен быть в git |
| не закоммичен `go.sum` | `go mod tidy` и commit |
