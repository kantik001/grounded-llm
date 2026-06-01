# RAG eval — quality regression by domain

| File | Domain | Questions | Language |
|------|--------|-----------|----------|
| `rag_default_baseline.jsonl` | `default` | 12 | RU (legacy demo) |
| `rag_default_en_baseline.jsonl` | `default` | 18 | EN (Phase A international) |

Line format:

```json
{
  "domain_id": "default",
  "question": "How many vacation days?",
  "expect_contains": ["28"],
  "expect_context": true,
  "category": "policy"
}
```

## Run

```bash
# Python RAG on :5000
export PYTHON_RAG_URL=http://localhost:5000/rag/context
python scripts/run_rag_eval.py --suite default_en
python scripts/run_rag_eval.py --suite all
make eval-retrieval
```

Reports: `eval/results/`.

CI validates JSONL structure via `tests/test_eval_baseline.py` (EN suite requires ≥15 cases).
