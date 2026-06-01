# HR Domain Pack (English)

**SKU:** Policy Assistant — HR & Employee Handbook  
**Domain ID:** `default` (demo) or client-specific `hr`  
**Locale:** `config/locales/en/`

---

## What you sell

A ready-to-deploy **internal HR Q&A assistant** grounded in company policies:

- Paid leave and vacation planning  
- Sick leave reporting  
- Remote / hybrid work rules  
- Conduct and HR escalation  

**Pitch (one line):**  
*Employees get instant, cited answers from your HR handbook — hosted in your infrastructure.*

---

## Included in this pack

| Asset | Path |
|-------|------|
| Demo knowledge base (EN) | `data/default/*_en.txt` |
| RAG prompts | `config/locales/en/prompts.json` |
| Onboarding chips | `config/locales/en/onboarding.json` |
| UI branding | `config/locales/en/branding.json` |
| Few-shot retrieval hints | `config/locales/en/few_shot.json` |
| Eval baseline (EN) | `eval/rag_default_en_baseline.jsonl` |
| Demo script | [DEMO_SCRIPT.md](./DEMO_SCRIPT.md) |

---

## Client onboarding (2–5 days)

1. Copy domain entry in `config/domains.json` → `hr` (or rename `default`).
2. Replace demo TXT/PDF/DOCX with client policies under `data/{tenant}/hr/`.
3. Tune `config/locales/en/prompts.json` tone (formal / friendly).
4. Update onboarding questions to match client topics.
5. `python scripts/reindex_rag.py` or `POST /admin/reindex`.
6. Run `python scripts/run_rag_eval.py --suite default_en`.
7. Pilot per [PILOT_PLAYBOOK.md](../PILOT_PLAYBOOK.md).

---

## Pricing guide (reference)

| Deliverable | Guide |
|-------------|-------|
| Pack setup (docs + prompts + deploy) | Included in pilot or $3k–8k setup |
| Pilot (8 weeks) | $8k–25k |
| Annual license | $24k–80k |

---

## Out of scope (Phase A)

- Payroll calculation, individual PII lookup  
- Legal advice disclaimer remains in branding  
- SSO / RBAC — Phase B  

---

See also: [SECURITY_BRIEF.md](../SECURITY_BRIEF.md), [LOCALE_GUIDE.md](../LOCALE_GUIDE.md).
