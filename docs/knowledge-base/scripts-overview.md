# Разбор: папка `scripts/` / Scripts

| Файл | Задача |
|------|--------|
| `reindex_rag.py` | Пересборка Chroma из `data/` |
| `run_rag_eval.py` | Прогон `eval/*.jsonl`, отчёт в `eval/results/` |
| `smoke.sh` / `smoke.ps1` | Smoke API Go |
| `create_github_repo.ps1` | Публикация на GitHub |

---

## `reindex_rag.py`

После изменений в `data/{domain_id}/` (`.txt`, `.pdf`, `.docx`):

```bash
python scripts/reindex_rag.py
```

Устанавливает `FORCE_RAG_REINDEX=true`, вызывает `create_vector_store()`.

**Зависимости:** `pip install -r api/requirements.txt`

**Альтернативы:** admin UI reindex, `POST /admin/reindex`, `FORCE_RAG_REINDEX=true` при старте `python`.

---

## `run_rag_eval.py`

```bash
set PYTHON_RAG_URL=http://localhost:5000/rag/context
python scripts/run_rag_eval.py --suite default
python scripts/run_rag_eval.py --suite all
```

Suites: см. `SUITES` в скрипте (`default` → `eval/rag_default_baseline.jsonl`).

Режим `--full` (опционально): end-to-end через Go `/chat` — нужен `LLM_API_KEY`.

---

## Smoke

```bash
make smoke
# TELEGRAM_AUTH_DISABLED=true, server :8080
```

---

## Makefile

`make test`, `make reindex`, `make eval-retrieval`, `make up-build` — см. `Makefile`.

Compose project: **`grounded_llm`**.
