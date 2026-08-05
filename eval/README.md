# RAG eval — quality regression by domain

Retrieval baselines live here (not under `tests/`): CI gates product quality.

| File | Domain | Questions | Language / notes |
|------|--------|-----------|------------------|
| `rag_default_baseline.jsonl` | `default` | 12 | **RU** — matches `*_policy_ru.txt`; keep for RU KB coverage |
| `rag_default_en_baseline.jsonl` | `default` | 21 | **EN** — primary HR demo / Phase A gate |
| `rag_it_support_baseline.jsonl` | `it_support` | 16 | EN (IT support template) |
| `rag_legal_faq_baseline.jsonl` | `legal_faq` | 13 | EN (Legal FAQ template) |
| `rag_adversarial_baseline.jsonl` | mixed | 30 | EN (retrieval adversarial) |
| `rag_hybrid_baseline.jsonl` | mixed | 7 | EN (keyword-heavy; only when `RAG_RETRIEVAL_MODE=hybrid`) |
| **Retrieval total** | — | **99** | All JSONL suites above |
| `rag_adversarial_e2e.jsonl` | mixed | 5 | EN (adversarial `/message` E2E; not in 99) |

Schemas (CI via `tests/test_eval_baseline.py`):

- [`schemas/baseline_case.schema.json`](./schemas/baseline_case.schema.json)
- [`schemas/adversarial_e2e_case.schema.json`](./schemas/adversarial_e2e_case.schema.json)

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
python scripts/run_rag_eval.py --suite default          # RU suite
python scripts/run_rag_eval.py --suite it_support
python scripts/run_rag_eval.py --suite adversarial
python scripts/run_rag_eval.py --suite all
python scripts/run_adversarial_e2e.py --base-url http://localhost:8080
make eval-retrieval
```

## CI gates

| Job | What it checks |
|-----|----------------|
| `eval-baseline-validate` | JSONL + JSON Schema (fast, no RAG) |
| `eval-retrieval-gate` | Reindex → start Python → **99 retrieval cases** must pass (hybrid suite when `RAG_RETRIEVAL_MODE=hybrid`) |

Local equivalent of the retrieval gate:

```bash
pip install -r api/requirements.txt requests
make eval-retrieval-ci
# or: bash scripts/ci_eval_retrieval.sh
```

Reports: `eval/results/` (gitignored).
