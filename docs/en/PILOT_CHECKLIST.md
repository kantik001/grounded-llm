# Pilot checklist — production readiness

Use with [PILOT_OFFER.md](./PILOT_OFFER.md). Check before customer UAT.

## Day 0

- [ ] `python scripts/pilot_day0.py --pack hr|it_support --tenant <id>`
- [ ] Customer docs in printed `data_dir` (txt/pdf/docx)
- [ ] `python scripts/reindex_rag.py`
- [ ] Golden set: copy `eval/pilot_golden_template.jsonl` → `eval/rag_pilot_<tenant>.jsonl` and fill facts
- [ ] `python scripts/run_rag_eval.py --suite <suite>` green on in-scope items

## Security / prod knobs

- [ ] `VERIFY_FAITHFULNESS=enforce`
- [ ] Numeric verify on (default Spec path)
- [ ] API keys hashed; **tenant binding** enabled
- [ ] Membership / quotas as required
- [ ] `METRICS_TOKEN` set; `/metrics` not public
- [ ] CORS allowlist (no `*` in prod)
- [ ] `TELEGRAM_AUTH_DISABLED=false` (or OIDC) for real users
- [ ] Non-root images / Helm securityContext
- [ ] Prod compose **without** source bind mounts (`docker-compose.prod.yml`)

## Trust demo for stakeholders

- [ ] In-scope question → answer + source filename
- [ ] Out-of-scope (Moon/Mars) → refuse, no invented numbers
- [ ] Upload small policy → reindex → new question works

## Week 8 deliverables

- [ ] KPI report (Recall@k, citation %, verify rejects, refusals, latency)
- [ ] Runbook (up/down, reindex, key rotation)
- [ ] [SECURITY_BRIEF.md](./SECURITY_BRIEF.md) handed to IT
- [ ] Written go/no-go note
