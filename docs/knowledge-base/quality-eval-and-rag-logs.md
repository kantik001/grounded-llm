# Eval RAG и логи `[RAG]` / Quality eval & logs

**Скрипт / Script:** `scripts/run_rag_eval.py`  
**Наборы / Suites:** `eval/rag_*_baseline.jsonl`  
**Логи / Logs:** `server/rag_log.go` → `[RAG] domain_id=...`

---

## Eval retrieval

| Файл | Домен | Вопросов |
|------|-------|----------|
| `rag_default_baseline.jsonl` | `default` | 5 |

Формат строки:

```json
{
  "domain_id": "default",
  "question": "Сколько дней отпуска?",
  "expect_contains": ["28"],
  "expect_context": true
}
```

### Запуск

```bash
export PYTHON_RAG_URL=http://localhost:5000/rag/context
python scripts/run_rag_eval.py --suite default
make eval-retrieval
```

Отчёты: `eval/results/YYYYMMDD_HHMMSS.json`.

---

## Structured logs

Go пишет (без тела LLM):

```
[RAG] domain_id=default session_id=... fragments=4 verify_pass=true soft_fail=false reason="..." question="..."
```

Используйте для метрик: hit rate, verify pass rate, top failed questions.

---

## Когда гонять eval

- после reindex
- после смены `prompts.json` / chunking
- после добавления domain pack документов
- перед релизом domain pack

---

## End-to-end (--full)

Опционально через Go `/chat` — нужен `LLM_API_KEY`. По умолчанию CI **не** гоняет eval.

---

## Связанные docs

| Тема | Файл |
|------|------|
| Scripts | [scripts-overview.md](./scripts-overview.md) |
| Verify | [rag-verifier.md](./rag-verifier.md) |
| eval/ | [../eval/README.md](../eval/README.md) |
