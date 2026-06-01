# RAG eval — регрессии качества по домену

| Файл | Домен | Вопросов |
|------|-------|----------|
| `rag_default_baseline.jsonl` | `default` | 12 |

Формат строки:

```json
{
  "domain_id": "default",
  "question": "Сколько дней отпуска?",
  "expect_contains": ["28"],
  "expect_context": true,
  "category": "policy"
}
```

## Запуск

```bash
# Python RAG на :5000
export PYTHON_RAG_URL=http://localhost:5000/rag/context
python scripts/run_rag_eval.py --suite default
```

Отчёты: `eval/results/`.
