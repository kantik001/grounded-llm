# `rag/retrieval.py` — RAG retrieval

**Source:** `rag/retrieval.py`  
**Endpoint:** `POST /rag/context`  
**Next:** Go `server/rag_chat.go` → LLM

---

## Purpose

1. Accept question, `domain_id`, `tenant_id`, `locale`
2. `vector_store.search()` — fragments from Chroma
3. Build `context` + locale-specific `few_shot`
4. Return JSON for Go — **no LLM**

---

## `retrieve_rag_context(question, domain_id, tenant_id, locale)`

### Output fields

| Field | Purpose |
|-------|---------|
| `success` | context found |
| `error` | user-facing error message |
| `context` | text for prompt |
| `few_shot` | from `config/locales/{locale}/few_shot.json` |
| `category` | currently `"general"` |
| `fragments` | `{filename, content, excerpt, page?}` for verify + citations |
| `domain_id` | normalized id |

### Soft-fail errors

- empty question
- unknown `domain_id`
- `rag_enabled: false`
- no fragments in Chroma

---

## Few-shot

`few_shot_for(domain_id, category, locale)` — domain key in locale `few_shot.json`.

---

## Related files

| Topic | File |
|-------|------|
| Search | [rag-vector_store.md](./rag-vector_store.md) |
| Domains | [rag-domains_config.md](./rag-domains_config.md) |
| Go orchestration | [server-rag_chat.md](./server-rag_chat.md) |
