# Media checklist — dev.to platform article

Article draft: [from-vertical-rag-to-open-standard.md](./from-vertical-rag-to-open-standard.md)

Publish on dev.to as **Building Grounded LLM** series, part 2 (after horticulture passion post).

---

## Before recording

```bash
cp .env.example .env
# TELEGRAM_AUTH_DISABLED=true
# LLM_API_KEY=...  (or LLM_MOCK=true RAG_MOCK=true for UI-only GIFs)

docker compose up -d --build
python scripts/init_pack.py install hr   # or it_support
python scripts/reindex_rag.py
```

Use **English** UI: `DEFAULT_LOCALE=en` in `.env`.

---

## Assets to capture

| # | Type | Filename (suggested) | What to show | Duration |
|---|------|----------------------|--------------|----------|
| 0 | **Cover** | `cover-grounded-standard.png` | Spec doc + chat citations collage; Canva 1000×420 | static |
| 1 | **GIF** | `demo-chat-citations.gif` | 2–3 HR/IT questions, citations visible | 30–45s |
| 2 | **GIF** | `demo-conformance-cli.gif` | `python -m conformance spec` + `check --url` green | 20–30s |
| 3 | GIF (opt) | `demo-ci-eval-gate.gif` | GitHub Actions eval-retrieval-gate green | 15s |
| 4 | PNG | `screenshot-packs-registry.png` | `init_pack.py list` output | static |
| 5 | PNG | `screenshot-admin-upload.png` | admin.html upload + reindex | static |
| 6 | PNG (opt) | `screenshot-architecture.png` | README mermaid exported from GitHub or draw.io | static |

Store in repo (for GitHub rendering):

```
docs/assets/blog/platform-launch/
  cover-grounded-standard.png
  demo-chat-citations.gif
  demo-conformance-cli.gif
  ...
```

Or host on dev.to CDN only (upload when publishing).

---

## Recording tips (Windows)

- **Terminal GIF:** [ScreenToGif](https://www.screentogif.com/) or OBS → convert to GIF (≤8MB for dev.to)
- **Browser:** zoom 110%, hide bookmarks bar, use `localhost` chat
- **Conformance:** increase font size in terminal; dark theme contrasts well on dev.to

---

## dev.to publish settings

| Field | Value |
|-------|-------|
| **Title** | I built a grounded RAG assistant for my father's research papers — then turned it into an open platform I want to become an industry standard |
| **Tags** | `opensource`, `rag`, `ai`, `devops`, `standardization` |
| **Series** | Building Grounded LLM |
| **Canonical URL** | GitHub raw or `docs/en/blog/from-vertical-rag-to-open-standard.md` on main after merge |
| **Cover image** | `cover-grounded-standard.png` |

### Tags to add (discoverability, not @spam)

- `#machinelearning` `#llm` `#selfhosted` `#enterprise` `#googlecloud` (industry context only)

### @mentions — do / don't

| Do | Don't |
|----|-------|
| Link your [previous DEV post](https://dev.to/kantik001/my-father-wrote-the-papers-i-built-a-rag-assistant-so-growers-can-query-them-safely-1hi) | Cold-@ random Google employees |
| Ask for RFC feedback in comments | "Hey @Google promote my repo" |
| Compare **category** to NotebookLM / Vertex (intellectual) | Imply endorsement from Google |

### If you want Google-adjacent eyes (proper channels)

1. **dev.to tags** `#googlecloud` `#opensource` — organic discovery
2. **Google Open Source** — if you later post a conformance blog, submit to their interest form (not dev.to @)
3. **LinkedIn / X** — tag **products** (@GoogleCloud, @GoogleAI) only with a technical angle (spec + eval), not launch spam
4. **OSS communities** — LangChain Discord, r/selfhosted, HN Show — better ROI than executive @mentions

---

## Post-publish checklist

- [ ] Pin comment with links: repo, spec, conformance README, horticulture post
- [ ] Cross-post link on LinkedIn (Russian + EN audiences)
- [ ] Add article URL to `site/index.html` or README after public launch
- [ ] Reply to every comment in first 48h
