# Legal FAQ Template Pack (reference)

**Use case:** internal legal and compliance Q&A (NDA, privacy, contracts, IP, ethics)  
**Domain ID:** `legal_faq`  
**Locale:** `config/locales/en/`  
**Knowledge base:** `data/default/legal_faq/`

Third **official template** for Grounded LLM — deploy alongside [HR](./HR.md) and [IT Support](./IT_SUPPORT.md).

---

## What this template provides

Document-grounded answers for common legal/compliance questions:

- NDA duration and confidential information handling  
- GDPR / privacy contacts and retention  
- Contract signing thresholds and legal review SLAs  
- IP ownership and open-source policy  
- Whistleblower hotline and compliance training deadlines  

**One-liner:**

> Employees get cited answers from your legal/compliance policies — on your infrastructure, with measurable retrieval quality.

**Disclaimer:** This pack ships demo policy text. Replace with counsel-approved documents before production. The assistant cites policies; it does not replace legal advice.

---

## Included assets

| Asset | Path |
|-------|------|
| Demo knowledge base | `packs/legal_faq/data/*.txt` |
| Pack manifest | `packs/legal_faq/pack.yaml` |
| Eval baseline | `eval/rag_legal_faq_baseline.jsonl` (11 policy + 2 edge cases) |

---

## Deploy from template

```bash
python scripts/init_pack.py install legal_faq
python scripts/reindex_rag.py
python scripts/run_rag_eval.py --suite legal_faq
```

Optional hybrid retrieval (keyword rerank after vector search):

```bash
RAG_RETRIEVAL_MODE=hybrid python scripts/run_rag_eval.py --suite legal_faq
```

---

## Sample eval questions

| Question | Expected fact |
|----------|----------------|
| How long does the standard NDA last? | 2 years |
| DPO contact email? | privacy@company.com |
| Contract value requiring legal review? | USD 10,000 |
| Whistleblower hotline email? | ethics-hotline@company.com |

Full suite: `eval/rag_legal_faq_baseline.jsonl`.

---

## Related

- [packs/README.md](../../packs/README.md)
- [VECTOR_STORE.md](../VECTOR_STORE.md)
