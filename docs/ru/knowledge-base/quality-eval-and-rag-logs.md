# Eval RAG и логи `[RAG]`

**Скрипт:** `scripts/run_rag_eval.py`  
**Наборы:** `eval/rag_*_baseline.jsonl`  
**Логи:** `server/rag_log.go`

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

В CI может выполняться `eval-baseline-validate` для JSONL-файлов.

---

## Структурированные логи

Go пишет (без тела LLM):

```
[RAG] domain_id=default session_id=... fragments=4 verify_pass=true soft_fail=false reason="..." question="..."
```

Метрики: hit rate, verify pass rate, частые провалы.

---

## Когда гонять eval

- после reindex
- после смены промптов в `config/locales/`
- после добавления документов domain pack
- перед релизом domain pack

---

## End-to-end (`--full`)

Опционально через Go `POST /message` — нужен `LLM_API_KEY`. CI по умолчанию **не** гоняет полный LLM eval.

---

## Связанные статьи

| Тема | Файл |
|------|------|
| Скрипты | [scripts-overview.md](./scripts-overview.md) |
| Verify | [rag-verifier.md](./rag-verifier.md) |
| eval/ | [../../eval/README.md](../../eval/README.md) |
