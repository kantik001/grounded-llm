# Case Study — HR Policy Assistant (Pilot)

**Use case:** Internal HR Q&A over vacation, sick leave, remote work, and conduct policies.  
**Deployment:** Docker Compose on customer VPC (on-prem).  
**Domain pack:** HR template (`packs/hr`, `data/default/*_en.txt`).

> **Note:** Metrics below reflect a **structured pilot demo** with scripted questions and eval baselines — a template for real customer pilots, not a published third-party production deployment.

---

## Problem

HR teams spend hours answering repeat questions in chat and email. Public LLM tools are blocked by IT because answers are not cited and data may leave the network.

## Solution

Grounded LLM with:

- Documents-only answers with **filename citations**
- **Numeric verify** layer (e.g. “28 days” must appear in retrieved text)
- **Retrieval eval gate** in CI before each release
- Admin upload + async reindex for policy updates

## Pilot setup (5 days)

| Day | Activity |
|-----|----------|
| 1 | Deploy platform, configure SSO (optional), upload HR TXT/PDF policies |
| 2 | Install HR template pack, reindex, run eval suite |
| 3 | UAT with 8–10 HR champions (scripted + free-form questions) |
| 4 | Fix KB gaps from `kb_gaps` analytics, reindex |
| 5 | KPI readout + go/no-go for wider rollout |

## Sample KPIs (demo pilot, n=10 users, ~50 questions)

| Metric | Result | Target |
|--------|--------|--------|
| Retrieval eval pass (EN suite) | 18/18 | 100% |
| Answers with ≥1 citation | 46/50 (92%) | ≥90% |
| Verify pass rate (numeric answers) | 12/13 (92%) | ≥75% |
| Honest out-of-scope (off-topic questions) | 5/5 | 100% |
| Median time-to-answer (user perceived) | ~8 s | <15 s |
| Thumbs-up feedback | 34/50 (68%) | ≥60% |

## Representative questions

| Question | Outcome |
|----------|---------|
| How many paid vacation days? | **28** + citation `vacation_policy_en.txt` |
| CEO salary on the Moon? | Out-of-scope — no hallucination |
| Remote days per week? | **2** + citation `remote_work_policy_en.txt` |

## What IT cared about

- Data stays in customer Postgres + Chroma + `data/` volumes ([SECURITY_BRIEF.md](./SECURITY_BRIEF.md))
- Audit log for upload/reindex/admin login
- `/metrics` restricted to internal network

## Reusable artifacts

- [HR demo script](./domain-packs/DEMO_SCRIPT.md)
- [Analytics guide](./ANALYTICS_GUIDE.md)
- Eval suite: `eval/rag_default_en_baseline.jsonl`

## Next steps after pilot

1. Add legal/compliance domain pack  
2. Enable OIDC for admin + API keys for intranet embed  
3. Track verify pass rate weekly in admin analytics  

---

See also: [PILOT_PLAYBOOK.md](./PILOT_PLAYBOOK.md) · [PLATFORM_VISION.md](../../PLATFORM_VISION.md)
