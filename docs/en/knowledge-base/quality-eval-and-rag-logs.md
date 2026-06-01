# RAG eval and `[RAG]` logs

**Script:** `scripts/run_rag_eval.py`  
**Suites:** `eval/rag_*_baseline.jsonl`  
**Logs:** `server/rag_log.go` → `[RAG] domain_id=...`

---

## Retrieval eval

| File | Domain | Questions |
|------|--------|-----------|
| `rag_default_baseline.jsonl` | `default` | 5 |

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
python scripts/run_rag_eval.py --suite default
make eval-retrieval
```

Reports: `eval/results/YYYYMMDD_HHMMSS.json`.

CI may run `eval-baseline-validate` on baseline JSONL files.

---

## Structured logs

Go writes (no LLM body):

```
[RAG] domain_id=default session_id=... fragments=4 verify_pass=true soft_fail=false reason="..." question="..."
```

Use for metrics: hit rate, verify pass rate, top failed questions.

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
| eval/ | [../../eval/README.md](../../eval/README.md) |
