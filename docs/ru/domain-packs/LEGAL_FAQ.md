**Канон (EN):** [LEGAL_FAQ.md](../en/domain-packs/LEGAL_FAQ.md)

# Legal FAQ Template Pack (референс)

**Сценарий:** внутренний legal / compliance Q&A (NDA, privacy, контракты, IP, ethics)  
**Domain ID:** `legal_faq`  
**Locale:** `config/locales/en/`  
**KB:** `data/default/legal_faq/` (после install из pack)

Третий **официальный шаблон** — рядом с [HR](./HR.md) и [IT Support](./IT_SUPPORT.md).

---

## Что даёт шаблон

Цитируемые ответы из legal/compliance политик:

- Срок NDA и работа с конфиденциальной информацией  
- GDPR / privacy-контакты и retention  
- Пороги подписания контрактов и SLA legal review  
- IP и open-source policy  
- Whistleblower hotline и дедлайны compliance training  

**One-liner:**

> Сотрудники получают ответы с цитатами из legal/compliance политик — on-prem, с измеримым retrieval.

**Disclaimer:** в pack — демо-тексты. Перед prod замените документами, согласованными с counsel. Ассистент цитирует политики; **не** заменяет юридическую консультацию.

---

## Состав

| Актив | Путь |
|-------|------|
| Demo KB | `packs/legal_faq/data/*.txt` |
| Manifest | `packs/legal_faq/pack.yaml` |
| Eval | `eval/rag_legal_faq_baseline.jsonl` (**13** cases: 11 policy + 2 edge) |

---

## Деплой

```bash
python scripts/init_pack.py install legal_faq
python scripts/reindex_rag.py
python scripts/run_rag_eval.py --suite legal_faq
```

Опционально hybrid:

```bash
RAG_RETRIEVAL_MODE=hybrid python scripts/run_rag_eval.py --suite legal_faq
```

---

## Примеры eval

| Вопрос | Факт |
|--------|------|
| How long does the standard NDA last? | 2 years |
| DPO contact email? | privacy@company.com |
| Contract value requiring legal review? | USD 10,000 |
| Whistleblower hotline email? | ethics-hotline@company.com |

Полный suite: `eval/rag_legal_faq_baseline.jsonl`.

---

## См. также

- [packs/README.md](../../../packs/README.md)  
- [VECTOR_STORE.md](../VECTOR_STORE.md)  
- [BENCHMARK.md](../BENCHMARK.md)  
