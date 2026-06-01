# Pilot Playbook — Grounded LLM

**Package:** Policy Assistant Pilot (8 weeks)  
**Target buyer:** HR / IT sponsor at 200–5,000 employee organizations

---

## Offer summary

| Item | Included |
|------|----------|
| Scope | 1 knowledge domain (e.g. HR policies) |
| Documents | Up to 50 files or 200 MB (`.txt`, `.pdf`, `.docx`) |
| Users | Up to 500 active users |
| Deployment | On-prem or client private cloud (Docker) |
| Channels | Web chat (+ Telegram optional) |
| Languages | English (default locale bundles) |
| Support | 2 sync calls/week during pilot |

**Pilot fee (guide):** $8,000 – $25,000 USD depending on company size and on-site needs.

---

## Timeline

| Week | Activities |
|------|------------|
| **1** | Kickoff, security questionnaire, document inventory |
| **2** | Deploy to client environment, network + LLM API setup |
| **3** | Ingest documents, reindex, prompt tuning |
| **4** | UAT with HR team (10–20 testers) |
| **5** | Soft launch — 50–100 employees |
| **6** | Expand to 200–500 users, collect feedback |
| **7** | Tune failed questions, add missing docs |
| **8** | Final report + go/no-go for annual license |

**Demo-ready:** end of week 3.

---

## KPI (pilot success)

| Metric | Target |
|--------|--------|
| In-scope answers with citations | ≥85% |
| Verify pass rate (facts/numbers) | ≥75% |
| User thumbs-up on assistant messages | ≥70% |
| P95 response time (non-streaming) | <15 s |

Track honestly: **out-of-scope** (“not in knowledge base”) is expected and correct behavior.

---

## Statement of Work (template)

```text
STATEMENT OF WORK — Grounded LLM Policy Assistant Pilot

Client:     [Company name]
Provider:   [Your legal entity]
Duration:   8 weeks from kickoff
Fee:        USD [amount] — 50% on signature, 50% on Week 8 delivery

Scope:
- Deploy Grounded LLM in Client's environment
- Configure one (1) domain: HR / Employee Policies
- Ingest up to fifty (50) documents provided by Client
- Web chat for up to five hundred (500) users
- Admin training session (2 hours)
- Pilot closure report with agreed KPIs

Client responsibilities:
- Provide VM/hosting, network egress to LLM API (or local LLM)
- Provide documents and HR subject-matter contact
- Designate 10–20 UAT users

Provider responsibilities:
- Installation, configuration, prompt tuning
- Weekly status calls
- Bug fixes during pilot period

Exclusions:
- SSO, custom integrations, 24/7 SLA, content writing

Success criteria:
- System available ≥99% during business hours (Client infra excepted)
- KPI report delivered Week 8
- Go/no-go meeting for annual license

Confidentiality: mutual NDA / DPA as required
```

---

## 30-minute demo script

1. **Problem (3 min)** — Employees ask HR repeat questions; public ChatGPT is not allowed for internal policies.
2. **Live chat (10 min)** — Use [HR demo questions](./domain-packs/DEMO_SCRIPT.md).
3. **Trust (5 min)** — Citations, verify, honest “not found”, data stays on-prem.
4. **Admin (5 min)** — Upload PDF → reindex → new question works.
5. **Enterprise path (5 min)** — API keys, OpenAPI, multi-tenant, roadmap (RBAC, audit).
6. **Close (2 min)** — 8-week pilot, fixed fee, annual license option.

---

## Deliverables checklist

- [ ] Deployed stack (postgres, server, python, webapp)
- [ ] Admin credentials documented
- [ ] Reindex runbook
- [ ] Locale bundles tuned for client tone
- [ ] Eval baseline run (retrieval metrics)
- [ ] Week 8 PDF/Markdown report with KPIs and recommendations

---

## After pilot — annual license (anchor)

| Tier | Employees | Guide price / year |
|------|-----------|-------------------|
| Starter | 200–500 | $24k – $36k |
| Standard | 500–2,000 | $48k – $72k |
| Enterprise | 2,000+ | $80k – $150k+ |

Includes software updates and email support; implementation beyond pilot is separate SOW.

---

See also: [SECURITY_BRIEF.md](./SECURITY_BRIEF.md), [domain-packs/HR.md](./domain-packs/HR.md), [API_EXAMPLES.md](./API_EXAMPLES.md).
