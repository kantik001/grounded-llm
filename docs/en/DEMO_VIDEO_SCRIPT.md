# Demo video script (3–5 minutes)

Use this script to record a Loom/YouTube demo. Pair with [DEMO_SCRIPT.md](./domain-packs/DEMO_SCRIPT.md) for live Q&A.

**Setup before recording:** `docker compose up -d`, `python scripts/reindex_rag.py`, `TELEGRAM_AUTH_DISABLED=true`, open `http://localhost/?locale=en`.

---

## 0:00 — Hook (20 s)

> “Your company can’t paste internal policies into ChatGPT. Grounded LLM is an open platform for **cited, verified** document assistants — on-prem, with measurable retrieval quality.”

Show: README architecture diagram or running web chat.

---

## 0:20 — Problem (30 s)

> “HR and IT teams answer the same questions every week. You need answers **only from your documents**, with **sources**, inside **your** network.”

---

## 0:50 — Live chat (90 s)

Ask in web UI:

1. “How many paid vacation days do employees get?” → highlight **28** and **Sources** block  
2. “What is the CEO salary on the Moon?” → highlight honest **not in knowledge base** behavior  

---

## 2:20 — Trust (45 s)

> “Every answer links to a file. Numbers are **verified** against retrieved text — invented deadlines get flagged.”

Optional: show verify warning if you have a failing example staged.

---

## 3:05 — Integrator API (45 s)

Terminal:

```bash
pip install -e sdk/python
grounded-llm chat "How many vacation days?" --domain default
```

Or show `examples/python/chat_basic.py`.

---

## 3:50 — Quality gate (30 s)

> “Retrieval regressions are caught in CI — 46 eval cases across EN, RU, and IT packs.”

Show: GitHub Actions `eval-retrieval-gate` green (screenshot or browser).

---

## 4:20 — Close (20 s)

> “Template packs ship HR and IT assistants in days. MIT licensed, Docker-first, security brief for IT. Links in the repo.”

End card: GitHub URL, docs/en/QUICKSTART_SDK.md, SECURITY_BRIEF.md.

---

## Recording tips

- 1080p browser window, zoom 125% for readability  
- Mute Docker logs; use clean `.env` without real API keys on screen  
- Captions: enable YouTube auto-captions + manual fix for “RAG”, “Chroma”, “verify”  
