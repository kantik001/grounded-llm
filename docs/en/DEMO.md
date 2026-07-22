# 5-minute demo (HR pack)

Clone → up → ask. Shows **cited answers** from the default HR knowledge base and honest refusal when the answer is not in the docs.

For a longer sales walkthrough, see [domain-packs/DEMO_SCRIPT.md](./domain-packs/DEMO_SCRIPT.md).

---

## 1. Start

```bash
git clone https://github.com/kantik001/grounded-llm.git
cd grounded-llm
cp .env.example .env
```

In `.env` set at least:

```env
TELEGRAM_AUTH_DISABLED=true
LLM_API_KEY=sk-...          # OpenAI-compatible
# Optional local smoke without a real LLM:
# LLM_MOCK=true
# RAG_MOCK=true
```

```bash
docker compose up -d --build
python scripts/reindex_rag.py
```

Open **http://localhost/** (web chat) or hit the API at `http://localhost:8080`.

| Check | Expect |
|-------|--------|
| `curl -sf http://localhost:8080/health` | `200` |
| Web UI loads | Chat for domain `default` |

---

## 2. Ask (in-scope)

Paste these in the web chat. Each answer should include a **Sources** block with a policy filename.

| # | Question | Expect in answer |
|---|----------|------------------|
| 1 | How many paid vacation days do employees get? | **28** calendar days |
| 2 | How far in advance must vacation be planned in HR Portal? | **14** days |
| 3 | By what time must I notify my manager on the first sick day? | **10:00** |
| 4 | How many remote work days per week are allowed? | **2** days |
| 5 | Which days are recommended in office for team sync? | **Tuesday** and **Thursday** |
| 6 | What are core messenger availability hours? | **09:00–18:00** |

Same cases live in CI as suite `default_en` (`packs/hr/eval.jsonl`).

---

## 3. Ask (out-of-scope — trust)

| Question | Expect |
|----------|--------|
| What is the CEO salary on the Moon in 2099? | No invented salary; empty / “not in knowledge base” style refusal |
| How many vacation days for employees on Mars? | Same — no hallucination |

---

## 4. API one-liner (optional)

With `TELEGRAM_AUTH_DISABLED=true`:

```bash
# Session
curl -sS -X POST http://127.0.0.1:8080/api/session \
  -H "Content-Type: application/json" \
  -d '{"domain_id":"default"}'

# Message (replace SESSION_ID)
curl -sS -X POST http://127.0.0.1:8080/api/message \
  -H "Content-Type: application/json" \
  -d '{"session_id":"SESSION_ID","domain_id":"default","text":"How many paid vacation days do employees get?"}'
```

Or: `grounded-llm chat "How many vacation days?" --domain default` after `pip install -e "sdk/python[dev]"` — [QUICKSTART_SDK.md](./QUICKSTART_SDK.md).

---

## 5. If something fails

| Symptom | Fix |
|---------|-----|
| Empty / vague answers | Re-run `python scripts/reindex_rag.py`; check `docker compose logs python` |
| 401 in browser | `TELEGRAM_AUTH_DISABLED=true` and restart server |
| LLM errors | Set a valid `LLM_API_KEY`, or `LLM_MOCK=true` for structural smoke only |
| Wrong language UI | `DEFAULT_LOCALE=en` or `?locale=en` |

---

## Related

- [README Quick start](../../README.md#quick-start)
- [HR pack](./domain-packs/HR.md) · [30-min demo script](./domain-packs/DEMO_SCRIPT.md)
- [Benchmark / eval](./BENCHMARK.md)
