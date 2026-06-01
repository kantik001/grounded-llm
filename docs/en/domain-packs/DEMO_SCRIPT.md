# HR Demo Script (30 minutes)

Use the **default** domain with English locale (`?locale=en` or `DEFAULT_LOCALE=en`).

---

## Setup (before demo)

```bash
docker compose up -d
python scripts/reindex_rag.py
# Open http://localhost/ with TELEGRAM_AUTH_DISABLED=true on server for browser testing
```

Ensure eval passes (optional):

```bash
pip install requests
python scripts/run_rag_eval.py --suite default_en
```

---

## Opening (3 min)

> “Your employees ask HR the same questions every week. Public ChatGPT is not allowed for internal policies. Grounded LLM answers **only from your documents**, with sources, inside **your** infrastructure.”

---

## Live questions (10 min)

Ask in the web chat. Show **answer + Sources block** for each.

| # | Question | Expected fact in answer |
|---|----------|------------------------|
| 1 | How many paid vacation days do employees get? | 28 calendar days |
| 2 | How far in advance must vacation be planned in HR Portal? | 14 days |
| 3 | How many vacation days can be carried over to next year? | 14 days max |
| 4 | By what time must I notify my manager on the first sick day? | 10:00 |
| 5 | Within how many working days must I submit a sick note to HR? | 3 working days |
| 6 | How many remote days per week are allowed? | 2 days |
| 7 | Which days are recommended in office for team sync? | Tuesday and Thursday |
| 8 | What are core messenger availability hours? | 09:00–18:00 |

**Out-of-scope demo (trust):**

| Question | Expected behavior |
|----------|-------------------|
| What is the CEO salary on the Moon in 2099? | Honest “not in knowledge base” / empty context |
| How many vacation days on Mars? | Same — no hallucination |

---

## Trust segment (5 min)

- Point to **filename** in citation  
- Explain **verify** — invented numbers are blocked or softened  
- Show architecture: documents never used to train public models  

---

## Admin segment (5 min)

1. Open `/admin.html`  
2. Upload a small `.txt` policy snippet  
3. Click **Reindex RAG**  
4. Ask a question that only the new file answers  

---

## Close (2 min)

- 8-week pilot, fixed fee, KPI report  
- Annual license path  
- Hand [SECURITY_BRIEF.md](../SECURITY_BRIEF.md) to IT contact  

---

## Troubleshooting

| Issue | Check |
|-------|-------|
| Empty answers | Reindex; `docker compose logs python` |
| 401 in browser | `TELEGRAM_AUTH_DISABLED=true` |
| Wrong language UI | `DEFAULT_LOCALE=en` or `?locale=en` |
