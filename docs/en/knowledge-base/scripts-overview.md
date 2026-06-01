# `scripts/` folder

| File | Task |
|------|------|
| `reindex_rag.py` | Rebuild Chroma from `data/` |
| `run_rag_eval.py` | Run `eval/*.jsonl`, report to `eval/results/` |
| `init_domain.sh` / `init_domain.ps1` | Scaffold new domain + locale stubs |
| `smoke.sh` / `smoke.ps1` | Go API smoke test |
| `create_github_repo.ps1` | Publish to GitHub |

---

## `reindex_rag.py`

After changes in `data/{tenant}/{domain}/` (`.txt`, `.pdf`, `.docx`):

```bash
python scripts/reindex_rag.py
```

Sets `FORCE_RAG_REINDEX=true`, calls vector store rebuild.

**Dependencies:** `pip install -r api/requirements.txt`

**Alternatives:** admin UI reindex, `POST /admin/reindex`, `FORCE_RAG_REINDEX=true` on Python startup.

---

## `run_rag_eval.py`

```bash
set PYTHON_RAG_URL=http://localhost:5000/rag/context
python scripts/run_rag_eval.py --suite default
python scripts/run_rag_eval.py --suite all
```

Suites: see `SUITES` in script (`default` → `eval/rag_default_baseline.jsonl`).

Optional `--full`: end-to-end via Go message API — requires `LLM_API_KEY`.

---

## `init_domain.ps1` / `init_domain.sh`

Creates domain entry stubs, locale JSON templates, and `data/{tenant}/{domain}/` directory.

See `domain-pack-template/README.md`.

---

## Smoke

```bash
make smoke
# TELEGRAM_AUTH_DISABLED=true, server on :8080
```

---

## Makefile

`make test`, `make reindex`, `make eval-retrieval`, `make up-build` — see `Makefile`.

Compose project: **`grounded_llm`**.
