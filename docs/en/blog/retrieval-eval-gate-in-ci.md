# Why we run a retrieval eval gate in CI (not just unit tests)

*Most RAG projects test embeddings in notebooks. We test retrieval in every pull request.*

---

## The failure mode

RAG systems break quietly:

- A Chroma filter change returns HTTP 500 for multi-tenant queries  
- A new PDF loader strips numbers from chunks  
- Someone renames a policy file but forgets to update eval expectations  

Unit tests on pure functions do not catch this. Chat demos hide it until a customer asks about vacation days and gets silence.

## Our approach

Grounded LLM ships **JSONL eval suites** (`eval/*.jsonl`) with cases like:

```json
{"question": "How many paid vacation days?", "expect_contains": ["28"], "expect_context": ["28"]}
```

The script `scripts/run_rag_eval.py` hits the **Python RAG service only** — no LLM tokens burned. CI job `eval-retrieval-gate` reindexes Chroma and runs **all suites** (EN + RU + IT).

## Why retrieval-only first

| Layer | Cost to test | What it catches |
|-------|----------------|-----------------|
| Retrieval | CPU + embeddings | Missing chunks, bad filters, empty index |
| LLM generation | API $ + latency | Wording, hallucination style |
| Full E2E | Both | Everything — but too slow for every PR |

We added **LLM E2E nightly** (`eval-llm-nightly.yml`) for real API verification when `LLM_API_KEY` is configured. Daily merges stay fast; releases stay honest.

## Product implication

For HR/policy assistants, **wrong retrieval is worse than bland phrasing**. The eval gate encodes that priority.

## Try it locally

```bash
make eval-retrieval-ci
# or
python scripts/run_rag_eval.py --suite default_en
```

## Contribute eval cases

Adding a JSONL line when you fix a retrieval bug prevents the same regression twice. See [GOOD_FIRST_ISSUES.md](../../GOOD_FIRST_ISSUES.md).

---

[Read the eval README](../../eval/README.md) · [ROADMAP — Phase B](../ROADMAP.md)
