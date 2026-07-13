# Good first issues

Starter tasks for new contributors. Comment on an issue before opening a PR if you want to claim it.

Phases **1–11** are complete in the repo; this list focuses on docs, eval, packs, SDK, and polish. See [docs/en/ROADMAP.md](docs/en/ROADMAP.md) for strategic tracks after launch.

**Setup:** [CONTRIBUTING.md](CONTRIBUTING.md) · `make test` · `pip install -e "sdk/python[dev]"`

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
| 7 | Expand `eval/README.md` with examples for `expect_not_contains` and adversarial types | Markdown |

Run gate: `make eval-retrieval-ci`

---

## Template packs

| # | Task | Skills |
|---|------|--------|
| 8 | Add 3 eval cases to `packs/legal_faq/eval.jsonl` | JSONL |
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

## Ecosystem & polish

See [docs/en/ROADMAP.md](docs/en/ROADMAP.md) (post Phase 11).

| # | Task | Skills |
|---|------|--------|
| 16 | Document embed widget setup in `docs/ru/` (summary from [EMBED.md](docs/en/EMBED.md)) | Markdown |
| 17 | Add connector troubleshooting section to [CONNECTORS.md](docs/en/CONNECTORS.md) | Markdown |
| 18 | Scaffold a community pack with `python scripts/init_pack.py new` + sample eval | YAML + TXT |
| 19 | Add `conformance check --json` example to [conformance/README.md](conformance/README.md) | Markdown |

---

## How to pick

1. Choose **one** row  
2. Open a GitHub issue: “Claim: #5 — eval edge cases”  
3. PR with tests/eval pass if applicable  

Maintainers: create GitHub issues from this list and label `good first issue`.
