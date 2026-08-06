# Pilot offer — Grounded LLM (8 weeks)

One-page commercial/technical offer for an **on-prem document-grounded assistant** pilot.  
Reference stack: [grounded-llm](https://github.com/kantik001/grounded-llm) **v0.4.0+**.

Related: [PILOT_CHECKLIST.md](./PILOT_CHECKLIST.md) · [DEMO_SCRIPT.md](./domain-packs/DEMO_SCRIPT.md) · [SECURITY_BRIEF.md](./SECURITY_BRIEF.md)

---

## What you get

| Item | Detail |
|------|--------|
| **Scope** | 1 vertical (HR **or** IT support), 1 tenant, customer documents |
| **Deploy** | Docker Compose or Helm on customer infra (on-prem / VPC) |
| **Behavior** | Answers **only from KB**, with **citations**; invented numbers blocked (Spec verify); prod faithfulness `enforce` |
| **Duration** | **8 weeks** fixed |
| **Deliverables** | Running assistant · golden eval set · KPI report · runbook · security brief |

---

## Out of scope (pilot)

- Multi-region SaaS tenancy for many customers  
- Fine-tuning / training on customer data  
- Full SharePoint/Drive connector build (manual upload + reindex is default)  
- Agent / MCP workflows (optional demo only; see [grounded-agent](https://github.com/kantik001/grounded-agent))

---

## KPI (agreed at kickoff)

| KPI | Target (example — tune per customer) |
|-----|--------------------------------------|
| Retrieval **Recall@k** on golden set | ≥ 90% |
| Answers with **citation** | ≥ 95% of in-scope answers |
| **Verify reject** on adversarial / invented numbers | Caught (no silent pass) |
| Honest **refusal** on out-of-KB questions | Documented pass rate |
| p95 latency (answer) | Agreed SLA (e.g. &lt; 5s with cloud LLM) |

Golden questions: start from pack eval or `eval/pilot_golden_template.jsonl`.

---

## Timeline

| Week | Focus |
|------|--------|
| 0 | Kickoff, access, `scripts/pilot_day0.py`, sample docs |
| 1–2 | Ingest + reindex + golden set (20–50 Q) |
| 3–4 | Prompt/pack tune, verify modes, admin upload path |
| 5–6 | UAT with 5–10 internal users, adversarial checks |
| 7 | KPI freeze, hardening (auth, metrics token, CORS) |
| 8 | Report + go/no-go for annual license |

---

## Commercial frame (fill in)

- **Pilot fee:** ________ (fixed)  
- **Success → annual:** support / license path ________  
- **Data:** stays on customer infra; not used to train public models  

---

## Day-0 command (engineer)

```bash
# from grounded-llm root
python scripts/pilot_day0.py --pack hr --tenant pilot-acme
# or: --pack it_support
```

Then place customer PDFs/TXT under the printed `data_dir`, reindex, run eval.

---

## Contact

Kantemir Satibalov · [github.com/kantik001/grounded-llm](https://github.com/kantik001/grounded-llm)
