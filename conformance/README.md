# Conformance test suite

Verify that a **Grounded LLM deployment** (or alternative implementation) matches the published API contract and retrieval quality baselines.

Third-party implementors can run this suite against their build to claim **API v1 compatibility**.

---

## Quick run (reference implementation)

```bash
# Terminal 1 — server with mocks (no LLM/RAG bill)
export TELEGRAM_AUTH_DISABLED=true LLM_MOCK=true RAG_MOCK=true
export DATABASE_URL=postgres://grounded:grounded@localhost:5432/grounded?sslmode=disable
cd server && go run .

# Terminal 2 — conformance
pip install -r conformance/requirements.txt
export CONFORMANCE_BASE_URL=http://127.0.0.1:8080
pytest conformance/ -v
```

---

## What is tested

| Module | Checks |
|--------|--------|
| `test_openapi_http.py` | Public paths from OpenAPI return expected HTTP codes |
| `test_openapi_spec.py` | OpenAPI file validates (same as CI `openapi-validate`) |
| `test_golden_retrieval.py` | Optional — runs eval suites when `CONFORMANCE_RAG_URL` set |

---

## Environment variables

| Variable | Required | Description |
|----------|----------|-------------|
| `CONFORMANCE_BASE_URL` | for HTTP tests | Go server base, e.g. `http://127.0.0.1:8080` |
| `CONFORMANCE_RAG_URL` | optional | `http://127.0.0.1:5000/rag/context` for golden retrieval |
| `CONFORMANCE_SKIP_HTTP` | optional | Set `1` to skip live server tests |

---

## Golden retrieval

When Python RAG is running with indexed data:

```bash
export CONFORMANCE_RAG_URL=http://127.0.0.1:5000/rag/context
pytest conformance/test_golden_retrieval.py -v
```

This wraps `scripts/run_rag_eval.py --suite all` and fails if pass rate &lt; 100%.

---

## CI integration (Phase 4)

Planned job `conformance` in `.github/workflows/ci.yml`:

1. Start Postgres service
2. Start Go server (`LLM_MOCK`, `RAG_MOCK`)
3. `pytest conformance/ -v`

---

## Related

- [API_DEPRECATION_POLICY.md](../docs/en/API_DEPRECATION_POLICY.md)
- [COMPATIBILITY.md](../docs/en/COMPATIBILITY.md)
- OpenAPI: `/api/v1/openapi.json`
