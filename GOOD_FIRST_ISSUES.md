# Good first issues

Starter tasks for new contributors. Comment on an issue before opening a PR if you want to claim it.

**Setup:** [CONTRIBUTING.md](../CONTRIBUTING.md) · `make test` · `pip install -e "sdk/python[dev]"`

---

## Documentation (good first PR)

| # | Task | Skills |
|---|------|--------|
| 1 | Fix broken links in `docs/ru/` pointing to old paths | Markdown |
| 2 | Add SDK example: streaming chat with error handling | Python |
| 3 | Translate `QUICKSTART_SDK.md` summary into `docs/ru/` | EN/RU |
| 4 | Add screenshot placeholders to `DEMO_VIDEO_SCRIPT.md` | Docs |

---

## Eval baselines (high impact, low risk)

| # | Task | Skills |
|---|------|--------|
| 5 | Add 3 edge cases to `eval/rag_default_en_baseline.jsonl` (out-of-scope) | JSONL |
| 6 | Add IT support case for password reset doc | JSONL + IT pack |
| 7 | Document how to author eval cases in `eval/README.md` | Markdown |

Run gate: `make eval-retrieval-ci`

---

## Template packs

| # | Task | Skills |
|---|------|--------|
| 8 | Scaffold **Legal FAQ** pack (`packs/legal_faq/`) from `domain-pack-template/` | YAML + sample TXT |
| 9 | Add German locale stub in `config/locales/de/` (branding only) | JSON |
| 10 | Improve `packs/README.md` with install troubleshooting | Markdown |

---

## SDK / CLI

| # | Task | Skills |
|---|------|--------|
| 11 | Add `GroundedClient.admin_status()` wrapper (Basic Auth) | Python |
| 12 | CLI: `grounded-llm history --session ID` | Python |
| 13 | Increase SDK test coverage for streaming SSE parser | pytest + responses |

---

## Web / UX (small)

| # | Task | Skills |
|---|------|--------|
| 14 | Show verify warning badge in webapp when answer failed verify | JS |
| 15 | Admin analytics: export kb_gaps as CSV | JS |

---

## Phase 5 (standard publication)

See [docs/en/PHASE_5.md](docs/en/PHASE_5.md).

| # | Task | Skills |
|---|------|--------|
| 21 | Add legal FAQ template pack + eval JSONL | YAML + docs |
| 22 | Vector store adapter interface (Chroma → Qdrant stub) | Python + Go |
| 23 | GitHub Pages landing from `docs/en/spec/` | Markdown |
| 24 | Expand conformance CLI `--json` output for integrators | Python |

---

## How to pick

1. Choose **one** row  
2. Open a GitHub issue: “Claim: #5 — eval edge cases”  
3. PR with tests/eval pass if applicable  

Maintainers: create GitHub issues from this list and label `good first issue`.
