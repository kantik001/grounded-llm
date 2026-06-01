# Папка `scripts/`

| Файл | Задача |
|------|--------|
| `reindex_rag.py` | Пересборка Chroma из `data/` |
| `run_rag_eval.py` | Прогон `eval/*.jsonl`, отчёт в `eval/results/` |
| `init_domain.sh` / `init_domain.ps1` | Шаблон нового домена + локали |
| `smoke.sh` / `smoke.ps1` | Smoke-тест Go API |
| `create_github_repo.ps1` | Публикация на GitHub |

---

## `reindex_rag.py`

После изменений в `data/{tenant}/{domain}/`:

```bash
python scripts/reindex_rag.py
```

**Зависимости:** `pip install -r api/requirements.txt`

**Альтернативы:** админка → reindex, `POST /admin/reindex`, `FORCE_RAG_REINDEX=true` при старте Python.

---

## `run_rag_eval.py`

```bash
set PYTHON_RAG_URL=http://localhost:5000/rag/context
python scripts/run_rag_eval.py --suite default
python scripts/run_rag_eval.py --suite all
```

Наборы: см. `SUITES` в скрипте (`default` → `eval/rag_default_baseline.jsonl`).

Режим `--full` (опционально): end-to-end через Go `POST /message` — нужен `LLM_API_KEY`.

---

## `init_domain.ps1` / `init_domain.sh`

Создаёт заготовки домена, JSON локалей и каталог `data/{tenant}/{domain}/`.

См. `domain-pack-template/README.md`.

---

## Smoke

```bash
make smoke
# TELEGRAM_AUTH_DISABLED=true, server на :8080
```

---

## Makefile

`make test`, `make reindex`, `make eval-retrieval`, `make up-build` — см. `Makefile`.

Имя проекта Compose: **`grounded_llm`**.
