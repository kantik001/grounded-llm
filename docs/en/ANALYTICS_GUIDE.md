# Analytics Guide — measuring assistant quality

For product owners and platform engineers running Grounded LLM in pilot or production.

Technical API reference: [config/ANALYTICS.md](../../config/ANALYTICS.md).

---

## Why analytics matter

Document-grounded assistants fail in predictable ways:

1. **Retrieval miss** — right doc exists but wrong chunk retrieved  
2. **Generation drift** — good context, bad phrasing or invented detail  
3. **Verify fail** — numeric hallucination despite good prose  
4. **KB gap** — question outside uploaded documents  

Grounded LLM records signals for each path so you can prioritize fixes (data vs retrieval vs prompts).

---

## Key metrics

| Metric | Where | Action if bad |
|--------|-------|----------------|
| **questions_total** | Admin analytics | Low → adoption issue; high → capacity planning |
| **rag.verify_pass_rate** | Admin analytics | <75% → review prompts, KB quality, or verify rules |
| **rag.soft_fail** | Admin analytics | High → missing docs or bad chunking; add content or eval cases |
| **kb_gaps** | Admin analytics | Repeated themes → new KB articles or FAQ pack |
| **feedback thumbs** | Admin analytics | Negative cluster → inspect verify fails + citations |
| **Retrieval eval pass** | CI / `run_rag_eval.py` | Regression before release — block deploy |

---

## Weekly review ritual (15 min)

1. Open **Admin → Analytics** (`webapp/admin.html`)  
2. Check **verify_pass_rate** trend (7-day window)  
3. Scan **kb_gaps** — top 5 unanswered themes  
4. Cross-check **feedback** negatives  
5. If retrieval regressions suspected → run `make eval-retrieval-ci`  

---

## Connecting metrics to product decisions

| Observation | Likely cause | Product action |
|-------------|--------------|----------------|
| Soft-fail on “VPN” questions | Missing IT doc | Install IT support pack or upload VPN policy |
| Verify fail on dates/amounts | LLM paraphrased numbers | Tighten prompts; add eval case |
| Good verify, bad UX wording | Generation only | Prompt tuning (not more retrieval) |
| Citations present but wrong file | Retrieval ranking | Add reranker experiment (roadmap); expand eval |

---

## Pilot KPI template

Copy into customer readouts ([CASE_STUDY_HR_PILOT.md](./CASE_STUDY_HR_PILOT.md)):

| KPI | Week 1 | Week 2 | Target |
|-----|--------|--------|--------|
| verify_pass_rate | | | ≥75% |
| soft_fail rate | | | ↓ week over week |
| Avg questions/day | | | ↑ adoption |
| Eval suite pass | | | 100% |

---

## Privacy

Question previews in KB gaps are truncated to 80 characters. Do not index documents containing secrets you would not show an HR manager in a preview line.

---

See also: [PILOT_PLAYBOOK.md](./PILOT_PLAYBOOK.md) · [SECURITY_BRIEF.md](./SECURITY_BRIEF.md)
