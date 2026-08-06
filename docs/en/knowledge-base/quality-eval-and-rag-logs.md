# RAG eval and `[RAG]` logs

**Script:** `scripts/run_rag_eval.py`  
**Suites:** `eval/rag_*_baseline.jsonl` — **99** retrieval cases total (see [eval/README.md](../../eval/README.md))  
**Logs:** `internal/rag/log.go` → `[RAG] domain_id=...`

---

## Retrieval eval

| File | Domain | Questions |
|------|--------|-----------|
| `rag_default_baseline.jsonl` | `default` | 12 (RU) |
| `rag_default_en_baseline.jsonl` | `default` | 21 (EN) |
| `rag_it_support_baseline.jsonl` | `it_support` | 16 |
| `rag_legal_faq_baseline.jsonl` | `legal_faq` | 13 |
| `rag_adversarial_baseline.jsonl` | mixed | 30 |
| `rag_hybrid_baseline.jsonl` | mixed | 7 (hybrid mode) |
| **Total** | — | **99** |

Line format:

```json
{
  "domain_id": "default",
  "question": "How many vacation days?",
  "expect_contains": ["28"],
  "expect_context": true
}
```

### Run

```bash
export PYTHON_RAG_URL=http://localhost:5000/rag/context
python scripts/run_rag_eval.py --suite default_en
make eval-retrieval
make eval-retrieval-ci
```

Reports: `eval/results/YYYYMMDD_HHMMSS.json`.

CI: `eval-baseline-validate` (JSONL schema) + `eval-retrieval-gate` (reindex + live RAG + all suites).

---

## Structured logs

Go writes (no LLM body):

```
[RAG] domain_id=default session_id=... fragments=4 verify_pass=true soft_fail=false reason="..." question="..."
```

Use for metrics: hit rate, verify pass rate, top failed questions. Product analytics: [ANALYTICS_GUIDE.md](../ANALYTICS_GUIDE.md).

---

## When to run eval

- after reindex
- after locale prompt / chunking changes
- after adding domain pack documents
- before domain pack release

---

## End-to-end (`--full`)

Optional via Go message API — requires `LLM_API_KEY`. CI does **not** run full LLM eval by default.

---

## Related docs

| Topic | File |
|-------|------|
| Scripts | [scripts-overview.md](./scripts-overview.md) |
| Verify | [rag-verifier.md](./rag-verifier.md) |
| Benchmark | [../BENCHMARK.md](../BENCHMARK.md) |
| eval/ | [../../eval/README.md](../../eval/README.md) |
