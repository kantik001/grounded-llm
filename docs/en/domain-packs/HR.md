# HR Template Pack (reference)

**Use case:** internal HR and employee handbook Q&A  
**Domain ID:** `default` (demo) or your slug e.g. `hr`  
**Locale:** `config/locales/en/`

This is the **reference template** for Grounded LLM. Copy and adapt it to ship a new document-grounded assistant in days.

---

## What this template provides

A ready-to-run **policy Q&A assistant** grounded in company documents:

- Paid leave and vacation planning  
- Sick leave reporting  
- Remote / hybrid work rules  
- Conduct and HR escalation  

**One-liner:**

> Employees get instant, cited answers from your handbook — deployed in your infrastructure.

---

## Included assets

| Asset | Path |
|-------|------|
| Demo knowledge base (EN) | `data/default/*_en.txt` |
| RAG prompts | `config/locales/en/prompts.json` |
| Onboarding chips | `config/locales/en/onboarding.json` |
| UI branding | `config/locales/en/branding.json` |
| Few-shot retrieval hints | `config/locales/en/few_shot.json` |
| Eval baseline (EN) | `eval/rag_default_en_baseline.jsonl` |
| Pack manifest | `packs/hr/pack.yaml` |
| Live demo script | [DEMO_SCRIPT.md](./DEMO_SCRIPT.md) |

---

## Deploy from template (2–5 days)

**Recommended — install from pack:**

```bash
python scripts/init_pack.py install hr
python scripts/reindex_rag.py
python scripts/run_rag_eval.py --suite default_en
```

**Manual steps:**

1. Scaffold: copy `default` domain entry in `config/domains.json` (or use pack install above).
2. Replace demo files with your policies under `data/{tenant}/hr/`.
3. Tune `config/locales/en/prompts.json` and `branding.json`.
4. Update onboarding questions for your topics.
5. `python scripts/reindex_rag.py` or `POST /admin/reindex`.
6. Validate: `python scripts/run_rag_eval.py --suite default_en`.
7. Review [SECURITY_BRIEF.md](../SECURITY_BRIEF.md) with your IT team before production.

---

## Out of scope (platform today)

- Payroll calculation, individual PII lookup  
- SSO / RBAC — see [roadmap Phase B](../ROADMAP.md)  
- Legal advice — disclaimer remains in branding bundles  

---

## Related

- [PLATFORM_VISION.md](../../../PLATFORM_VISION.md) — platform positioning  
- [LOCALE_GUIDE.md](../LOCALE_GUIDE.md) — add locales  
- [domain-pack-template/](../../../domain-pack-template/) — generic scaffold  
