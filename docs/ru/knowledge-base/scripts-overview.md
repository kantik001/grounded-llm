# Папка `scripts/`

Полный каталог: [`scripts/README.md`](../../../scripts/README.md). Makefile: `make help`.

| Область | Примеры |
|---------|---------|
| RAG / eval | `reindex_rag.py`, `run_rag_eval.py`, `ci_eval_retrieval.sh` |
| Packs | `init_pack.py`, `pack_registry.py`, `init_domain.*` |
| Smoke / CI | `smoke.*`, `load_smoke.*`, `ci_start_mock_server.sh` |
| Codegen | `gen_retriever_grpc.py` |

---

## `reindex_rag.py`

После изменений в `data/{tenant}/{domain}/`:

```bash
python scripts/reindex_rag.py
# или
make reindex
```

**Зависимости:** `pip install -r api/requirements.txt`

**Альтернативы:** админка → reindex, `POST /admin/reindex`, `FORCE_RAG_REINDEX=true` при старте Python.

---

## `run_rag_eval.py`

```bash
set PYTHON_RAG_URL=http://localhost:5000/rag/context
python scripts/run_rag_eval.py --suite default_en
python scripts/run_rag_eval.py --suite all
make eval-retrieval SUITE=default_en
make eval-retrieval-ci
```

Наборы: из `eval/rag_{suite}_baseline.jsonl` (`discover_suites()`). `--suite all` — все baseline.

Режим `--full` (опционально): E2E через Go `POST /message` — нужен `LLM_API_KEY`.

Базовые наборы: [`eval/README.md`](../../../eval/README.md).

---

## Packs / домены

```bash
python scripts/init_pack.py list
python scripts/init_pack.py install it_support
```

`init_domain.sh` / `init_domain.ps1` — каркас одного домена. Для шаблонов предпочтительнее packs CLI — [packs/README.md](../../../packs/README.md).

---

## Smoke / load

```bash
make smoke
make load-smoke
make backup-smoke
```

---

## Makefile

`make test`, `make reindex`, `make eval-retrieval`, `make up-build` — см. корневой `Makefile`.

Имя проекта Compose: **`grounded_llm`**.
