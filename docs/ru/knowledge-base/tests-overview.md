# Тесты

**Папки:** `tests/` (pytest), `server/*_test.go` (Go)  
**CI:** [github-ci.yml.md](./github-ci.yml.md)

---

## Python (`tests/`)

| Файл | Проверяет |
|------|-----------|
| `test_verifier.py` | `rag/verifier.py` — числа, дисклеймер |
| `test_domains_config.py` | `rag/domains_config.py`, `domains.json` |
| `test_document_loaders.py` | Форматы KB, загрузка `.txt` |

```bash
pip install -r tests/requirements-test.txt
pytest tests/ -v
```

Env: `DOMAINS_CONFIG_PATH=config/domains.json`

---

## Go (`server/`)

| Файл | Проверяет |
|------|-----------|
| `domains_test.go` | `normalizeDomainID` |
| `rag_chat_test.go` | verify, disclaimer, clean answer |
| `admin_test.go` | безопасное имя файла (txt/pdf/docx) |
| `auth_telegram_test.go` | initData HMAC |
| `locale_test.go` | нормализация locale, заголовок `X-Locale` |

```bash
cd server && go test -v ./...
```

---

## Не покрыто unit-тестами

- Chroma end-to-end
- LLM API
- Полный Docker compose (только ручной `make smoke`)

---

## Как в CI

```bash
make test
```
