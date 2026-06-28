# IT Support Template Pack (reference)

**Use case:** internal IT helpdesk Q&A (password, VPN, hardware, email, SLAs)  
**Domain ID:** `it_support`  
**Locale:** `config/locales/en/`  
**Knowledge base:** `data/default/it_support/`

Second **official template** for Grounded LLM — deploy alongside or instead of the [HR template](./HR.md).

---

## What this template provides

Document-grounded answers for common employee IT questions:

- Password reset and account lockout  
- VPN access requests  
- Laptop / hardware support  
- Phishing reports and email limits  
- IT Portal hours and ticket SLAs (P1–P3)

**One-liner:**

> Employees get cited answers from your IT runbooks — on your infrastructure, with measurable retrieval quality.

---

## Included assets

| Asset | Path |
|-------|------|
| Demo knowledge base | `data/default/it_support/*.txt` |
| Domain entry | `config/domains.json` → `it_support` |
| RAG prompts | `config/locales/en/prompts.json` → `it_support` |
| Onboarding chips | `config/locales/en/onboarding.json` → `it_support` |
| Few-shot hints | `config/locales/en/few_shot.json` → `it_support` |
| Eval baseline | `eval/rag_it_support_baseline.jsonl` (16 cases) |
| Pack manifest | `packs/it_support/pack.yaml` |

---

## Deploy from template (2–5 days)

**Recommended — install from pack:**

```bash
python scripts/init_pack.py install it_support
python scripts/reindex_rag.py
python scripts/run_rag_eval.py --suite it_support
```

**Manual steps:**

1. Copy or keep domain `it_support` in `config/domains.json`.
2. Replace demo TXT files with your runbooks under `data/{tenant}/it_support/`.
3. Tune `config/locales/en/prompts.json` → `it_support` (tone, disclaimers).
4. Update onboarding chips for your top tickets.
5. `python scripts/reindex_rag.py` or `POST /admin/reindex`.
6. Validate: `python scripts/run_rag_eval.py --suite it_support`.
7. Review [SECURITY_BRIEF.md](../SECURITY_BRIEF.md) with IT security before production.

Legacy scaffold (data dir only): `./scripts/init_domain.sh it_support default`

---

## Sample eval questions

| Question | Expected fact |
|----------|----------------|
| How long is a password reset link valid? | 24 hours |
| Which VPN client is approved? | GlobalProtect |
| P1 initial response time? | 1 hour |
| Where to report phishing? | security@company.com |

Full suite: `eval/rag_it_support_baseline.jsonl`.

---

## Out of scope

- Live ticket creation in ServiceNow/Jira (API integration — custom SOW)  
- Automated password reset execution  
- SSO configuration — see platform roadmap Phase B  

---

## Related

- [HR template](./HR.md)  
- [PLATFORM_VISION.md](../../../PLATFORM_VISION.md)  
- [domain-pack-template/](../../../domain-pack-template/)  
- [packs/README.md](../../../packs/README.md)  
