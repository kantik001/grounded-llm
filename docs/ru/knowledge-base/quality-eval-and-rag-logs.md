# RAG eval и логи `[RAG]`

**Скрипт:** `scripts/run_rag_eval.py`  
**Сьюиты:** `eval/rag_*_baseline.jsonl` — всего **99** retrieval-кейсов ([eval/README.md](../../../eval/README.md))  
**Логи:** `internal/rag/log.go` → `[RAG] domain_id=...`  
**EN:** [quality-eval-and-rag-logs.md](../../en/knowledge-base/quality-eval-and-rag-logs.md)

---

## Retrieval eval

| Файл | Domain | Вопросов |
|------|--------|----------|
| `rag_default_baseline.jsonl` | `default` | 12 (RU) |
| `rag_default_en_baseline.jsonl` | `default` | 21 (EN) |
| `rag_it_support_baseline.jsonl` | `it_support` | 16 |
| `rag_legal_faq_baseline.jsonl` | `legal_faq` | 13 |
| `rag_adversarial_baseline.jsonl` | mixed | 30 |
| `rag_hybrid_baseline.jsonl` | mixed | 7 (только hybrid) |
| **Итого** | — | **99** |

Формат строки:

```json
{
  "domain_id": "default",
  "question": "How many vacation days?",
  "expect_contains": ["28"],
  "expect_context": true
}
```

### Запуск

```bash
export PYTHON_RAG_URL=http://localhost:5000/rag/context
python scripts/run_rag_eval.py --suite default_en
make eval-retrieval
make eval-retrieval-ci
```

Отчёты: `eval/results/YYYYMMDD_HHMMSS.json`.

CI: `eval-baseline-validate` + `eval-retrieval-gate` (reindex + live RAG + все сьюиты).

---

## Структурированные логи

```
[RAG] domain_id=default session_id=... fragments=4 verify_pass=true soft_fail=false reason="..." question="..."
```

Метрики продукта: [../ANALYTICS_GUIDE.md](../ANALYTICS_GUIDE.md) · бенчмарк: [../BENCHMARK.md](../BENCHMARK.md).

---

## Когда гонять eval

- после reindex
- после смены промптов / chunking
- после добавления документов domain pack
- перед релизом пака

---

## Связанные документы

| Тема | Файл |
|------|------|
| Скрипты | [scripts-overview.md](./scripts-overview.md) |
| Verify | [rag-verifier.md](./rag-verifier.md) |
| CI | [github-ci.yml.md](./github-ci.yml.md) |
| eval/ | [../../../eval/README.md](../../../eval/README.md) |
