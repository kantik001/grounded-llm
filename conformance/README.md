# Conformance test suite

Verify that a **Grounded LLM deployment** (or alternative implementation) matches the published API contract and retrieval quality baselines.

Third-party implementors can run this suite to claim **Grounded-compatible** ([RFC-0001](../docs/en/rfcs/RFC-0001-grounded-compatible.md)).

---

## CLI (recommended)

```bash
pip install -r conformance/requirements.txt

python -m conformance spec
python -m conformance http --url http://127.0.0.1:8080
python -m conformance retrieval --rag-url http://127.0.0.1:5000/rag/context
python -m conformance check --url http://127.0.0.1:8080
python -m conformance all --url http://127.0.0.1:8080 --rag-url http://127.0.0.1:5000/rag/context

# JSON for integrator CI (stdout is JSON only)
python -m conformance spec --json
```

Makefile shortcuts: `make conformance-spec`, `make conformance-check URL=...`

---

## Quick run (reference implementation)

```bash
# Terminal 1 — server with mocks (no LLM/RAG bill)
export TELEGRAM_AUTH_DISABLED=true LLM_MOCK=true RAG_MOCK=true
export DATABASE_URL=postgres://grounded:grounded@localhost:5432/grounded?sslmode=disable
cd server && go run .

# Terminal 2 — conformance
pip install -r conformance/requirements.txt
python -m conformance check --url http://127.0.0.1:8080
```

---

## What is tested

| Module | Checks |
|--------|--------|
| `test_openapi_http.py` | Public paths from OpenAPI return expected HTTP codes |
| `test_openapi_spec.py` | OpenAPI file validates (same as CI `openapi-validate`) |
| `test_golden_retrieval.py` | Runs eval suites when `CONFORMANCE_RAG_URL` set |

---

## Environment variables

| Variable | Required | Description |
|----------|----------|-------------|
| `CONFORMANCE_BASE_URL` | for HTTP tests | Go server base, e.g. `http://127.0.0.1:8080` |
| `CONFORMANCE_RAG_URL` | optional | `http://127.0.0.1:5000/rag/context` for golden retrieval |
| `CONFORMANCE_SKIP_HTTP` | optional | Set `1` to skip live server tests |

---

## CI

| Job | Command |
|-----|---------|
| `conformance-spec` | `python -m conformance spec` |
| `smoke-api` | HTTP + adversarial E2E against running server |

---

## Related

- [GROUNDED_SPEC_v1.md](../docs/en/spec/GROUNDED_SPEC_v1.md)
- [API_DEPRECATION_POLICY.md](../docs/en/API_DEPRECATION_POLICY.md)
- [COMPATIBILITY.md](../docs/en/COMPATIBILITY.md)
- [BENCHMARK.md](../docs/en/BENCHMARK.md)
- OpenAPI: `/api/v1/openapi.json`
