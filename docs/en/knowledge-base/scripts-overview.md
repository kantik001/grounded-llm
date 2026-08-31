# `scripts/` folder

Full catalog: [`scripts/README.md`](../../../scripts/README.md). Makefile: `make help`.

| Area | Examples |
|------|----------|
| RAG / eval | `reindex_rag.py`, `backfill_kb_registry.py`, `run_rag_eval.py`, `ci_eval_retrieval.sh` |
| Packs | `init_pack.py`, `pack_registry.py`, `init_domain.*` |
| Smoke / CI | `smoke.*`, `load_smoke.*`, `ci_start_mock_server.sh` |
| Codegen | `gen_retriever_grpc.py` |

---

## `reindex_rag.py`

After new documents are registered (upload, connector, pack install):

```bash
python scripts/reindex_rag.py
# or
make reindex
```

Sets `FORCE_RAG_REINDEX=true`, rebuilds the configured vector backend (+ sparse when hybrid).

**Dependencies:** `pip install -r api/requirements.txt`

**Alternatives:** admin ingest, `POST /admin/reindex`, `FORCE_RAG_REINDEX=true` on Python startup.

---

## `backfill_kb_registry.py`

Register existing files from git-tracked `data/` into Postgres + blob store (one-time migration):

```bash
python scripts/backfill_kb_registry.py
python scripts/backfill_kb_registry.py --tenant default --domain default --dry-run
```

Then run `POST /admin/ingest`. See [KB_SOURCE_OF_TRUTH.md](../KB_SOURCE_OF_TRUTH.md).

---

## `run_rag_eval.py`

```bash
set PYTHON_RAG_URL=http://localhost:5000/rag/context
python scripts/run_rag_eval.py --suite default_en
python scripts/run_rag_eval.py --suite all
make eval-retrieval SUITE=default_en
make eval-retrieval-ci   # reindex + Python + all suites
```

Suites: discovered from `eval/rag_{suite}_baseline.jsonl` (see `discover_suites()`). Use `--suite all` for every baseline.

Optional `--full`: end-to-end via Go message API — requires `LLM_API_KEY`.

Eval baselines: [`eval/README.md`](../../../eval/README.md).

---

## Packs / domains

```bash
python scripts/init_pack.py list
python scripts/init_pack.py install it_support
# or make init-pack-list / make init-pack-install PACK=it_support
```

`init_domain.sh` / `init_domain.ps1` scaffold a single domain. Prefer packs CLI for templates — see [packs/README.md](../../../packs/README.md).

---

## Smoke / load

```bash
make smoke                 # TELEGRAM_AUTH_DISABLED=true, server on :8080
make load-smoke
make backup-smoke          # needs reachable Postgres
```

---

## Makefile

`make test`, `make reindex`, `make eval-retrieval`, `make up-build` — see root `Makefile`.

Compose project: **`grounded_llm`**.
