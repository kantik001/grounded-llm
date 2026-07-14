# RAG eval — quality regression by domain

| File | Domain | Questions | Language |
|------|--------|-----------|----------|
| `rag_default_baseline.jsonl` | `default` | 12 | RU (legacy demo) |
| `rag_default_en_baseline.jsonl` | `default` | 18 | EN (HR demo) |
| `rag_it_support_baseline.jsonl` | `it_support` | 16 | EN (IT support template) |
| `rag_hybrid_baseline.jsonl` | `default` | 5 | EN (keyword-heavy hybrid) |
| `rag_adversarial_e2e.jsonl` | mixed | 5 | EN (adversarial `/message` E2E) |

Line format (baseline):

```json
{
  "domain_id": "default",
  "question": "How many vacation days?",
  "expect_contains": ["28"],
  "expect_context": true,
  "category": "policy"
}
```

Adversarial cases add `adversarial_type` and optional `expect_not_contains` (Phase 4 runner):

```json
{
  "domain_id": "default",
  "question": "Do employees get 99 vacation days?",
  "adversarial_type": "wrong_number",
  "expect_contains": ["28"],
  "expect_not_contains": ["99"],
  "expect_context": true
}
```

Adversarial types: `wrong_number`, `missing_citation`, `cross_domain`, `prompt_injection`, `pii_trap`.

## Run locally

```bash
# Python RAG on :5000
export PYTHON_RAG_URL=http://localhost:5000/rag/context
python scripts/run_rag_eval.py --suite default_en
python scripts/run_rag_eval.py --suite it_support
python scripts/run_rag_eval.py --suite adversarial
python scripts/run_rag_eval.py --suite all
python scripts/run_adversarial_e2e.py --base-url http://localhost:8080
make eval-retrieval
```

## CI gates

| Job | What it checks |
|-----|----------------|
| `eval-baseline-validate` | JSONL structure (fast, no RAG) |
| `eval-retrieval-gate` | Reindex Chroma → start Python → **all suites must pass** (hybrid suite only when `RAG_RETRIEVAL_MODE=hybrid`) |

Local equivalent of the retrieval gate:

```bash
pip install -r api/requirements.txt requests
make eval-retrieval-ci
# or: bash scripts/ci_eval_retrieval.sh
```

Reports: `eval/results/`.
