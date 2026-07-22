# Grounded Benchmark

Public quality metrics for the reference implementation and **Grounded-compatible** claims ([RFC-0001](./rfcs/RFC-0001-grounded-compatible.md)).

---

## Suites

| Suite | File | Cases | Measures |
|-------|------|-------|----------|
| EN HR baseline | `eval/rag_default_en_baseline.jsonl` | 21 | Retrieval accuracy (+ paraphrase) |
| RU legacy | `eval/rag_default_baseline.jsonl` | 12 | Retrieval accuracy |
| IT support | `eval/rag_it_support_baseline.jsonl` | 16 | Cross-template retrieval |
| Adversarial | `eval/rag_adversarial_baseline.jsonl` | 30 | Wrong numbers, cross-domain, injection |
| Hybrid (keyword-heavy) | `eval/rag_hybrid_baseline.jsonl` | 7 | BM25+RRF regression (`RAG_RETRIEVAL_MODE=hybrid`) |
| Legal FAQ | `eval/rag_legal_faq_baseline.jsonl` | 13 | Cross-template retrieval |
| **Retrieval total** | — | **99** | All JSONL suites above (excl. E2E) |
| Adversarial E2E | `eval/rag_adversarial_e2e.jsonl` | 5 | Full `/message` path (mock or staging) |

---

## Run locally

```bash
# Retrieval only (Python RAG on :5000)
export PYTHON_RAG_URL=http://localhost:5000/rag/context
python scripts/run_rag_eval.py --suite all

# Summary JSON for README badge / release notes
python scripts/bench_report.py

# E2E adversarial (Go server with LLM_MOCK + RAG_MOCK)
python scripts/run_adversarial_e2e.py --base-url http://127.0.0.1:8080
```

---

## CI gates

| Job | Suites |
|-----|--------|
| `eval-retrieval-gate` | All retrieval JSONL (including adversarial) |
| `smoke-api` | Adversarial E2E (mock) |

---

## Release reporting

After each tag `v*.*.*`, maintainers SHOULD attach bench summary:

```bash
python scripts/bench_report.py --write eval/results/latest_bench.json
```

Example output:

```json
{
  "reference_impl": "grounded-llm",
  "version": "0.2.0",
  "suites": {
    "default_en": {"passed": 21, "total": 21, "pass_rate": 1.0},
    "it_support": {"passed": 16, "total": 16, "pass_rate": 1.0},
    "legal_faq": {"passed": 13, "total": 13, "pass_rate": 1.0},
    "adversarial": {"passed": 30, "total": 30, "pass_rate": 1.0},
    "hybrid": {"passed": 7, "total": 7, "pass_rate": 1.0}
  },
  "retrieval_total": {"passed": 99, "total": 99, "pass_rate": 1.0}
}
```

Future: publish leaderboard on GitHub Pages after public launch ([LAUNCH.md](docs/en/LAUNCH.md)).

---

## Related

- [eval/README.md](../../eval/README.md)
- [STANDARD_STRATEGY.md](./STANDARD_STRATEGY.md)
- [RFC-0001](./rfcs/RFC-0001-grounded-compatible.md)
