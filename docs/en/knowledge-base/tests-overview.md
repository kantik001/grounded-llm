# Tests overview

**Folders:** `tests/` (pytest), `server/*_test.go` (Go)  
**CI:** [github-ci.yml.md](./github-ci.yml.md)

---

## Python (`tests/`)

| File | Covers |
|------|--------|
| `test_verifier.py` | `rag/verifier.py` — numbers, disclaimer |
| `test_domains_config.py` | `rag/domains_config.py`, `config/domains.json` |
| `test_document_loaders.py` | KB formats, load `.txt` |

Dependencies: `tests/requirements-test.txt` (includes `langchain-community`, `pypdf`, `docx2txt`).

```bash
pip install -r tests/requirements-test.txt
pytest tests/ -v
```

Env: `DOMAINS_CONFIG_PATH=config/domains.json`

---

## Go (`server/`)

| File | Covers |
|------|--------|
| `domains_test.go` | `normalizeDomainID` |
| `rag_chat_test.go` | verify, disclaimer, clean answer |
| `admin_test.go` | safe filename (txt/pdf/docx) |
| `audit_test.go` | audit log query parsing, client IP, status check |
| `auth_telegram_test.go` | initData HMAC |
| `locale_test.go` | locale normalization, header resolution |

```bash
cd server && go test -v ./...
```

---

## Not covered by unit tests

- Chroma end-to-end
- LLM API
- Docker compose smoke (manual `make smoke` only)

---

## Local CI equivalent

```bash
make test
```
