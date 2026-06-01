# Разбор: тесты / Tests overview

**Папки / Folders:** `tests/` (pytest), `server/*_test.go` (Go)  
**CI:** [github-ci.yml.md](./github-ci.yml.md)

---

## Python (`tests/`)

| Файл | Что проверяет |
|------|---------------|
| `test_verifier.py` | `rag/verifier.py` — числа, disclaimer |
| `test_domains_config.py` | `rag/domains_config.py`, `config/domains.json` |
| `test_document_loaders.py` | форматы KB, load `.txt` |

Зависимости: `tests/requirements-test.txt` (включая `langchain-community`, `pypdf`, `docx2txt`).

```bash
pip install -r tests/requirements-test.txt
pytest tests/ -v
```

Env: `DOMAINS_CONFIG_PATH=config/domains.json`

---

## Go (`server/`)

| Файл | Что проверяет |
|------|---------------|
| `domains_test.go` | `normalizeDomainID` |
| `rag_chat_test.go` | verify, disclaimer, clean answer |
| `admin_test.go` | safe filename (txt/pdf/docx) |
| `auth_telegram_test.go` | initData HMAC |

```bash
cd server && go test -v ./...
```

---

## Что не покрыто unit-тестами

- Chroma end-to-end
- LLM API
- Docker compose smoke (только `make smoke` вручную)

---

## Локально как CI

```bash
make test
```
