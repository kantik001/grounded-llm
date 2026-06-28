# RAG eval — quality regression by domain

| File | Domain | Questions | Language |
|------|--------|-----------|----------|
| `rag_default_baseline.jsonl` | `default` | 12 | RU (legacy demo) |
| `rag_default_en_baseline.jsonl` | `default` | 18 | EN (HR demo) |
| `rag_it_support_baseline.jsonl` | `it_support` | 16 | EN (IT support template) |

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

## Run locally

```bash
# Python RAG on :5000
export PYTHON_RAG_URL=http://localhost:5000/rag/context
python scripts/run_rag_eval.py --suite default_en
python scripts/run_rag_eval.py --suite it_support
python scripts/run_rag_eval.py --suite all
make eval-retrieval
```

## CI gates

| Job | What it checks |
|-----|----------------|
| `eval-baseline-validate` | JSONL structure (fast, no RAG) |
| `eval-retrieval-gate` | Reindex Chroma → start Python → **all suites must pass** |

Local equivalent of the retrieval gate:

```bash
pip install -r api/requirements.txt requests
make eval-retrieval-ci
# or: bash scripts/ci_eval_retrieval.sh
```

Reports: `eval/results/`.
