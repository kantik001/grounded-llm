# Domain knowledge base

Runtime documents for RAG (not code). Supported: `.txt`, `.pdf`, `.docx` (UTF-8 for text).

## Layout (canonical)

```
data/
  {tenant_id}/
    {domain_id}/
      document.txt
```

| Tenant / domain | Path | Pack docs |
|-----------------|------|-----------|
| `default` / `default` (HR) | `data/default/default/` | [HR](../docs/en/domain-packs/HR.md) |
| `default` / `it_support` | `data/default/it_support/` | [IT Support](../docs/en/domain-packs/IT_SUPPORT.md) |
| `default` / `legal_faq` | `data/default/legal_faq/` | [Legal FAQ](../docs/en/domain-packs/LEGAL_FAQ.md) |

Official templates live under [`packs/*/data/`](../packs/); install copies into this tree (`python scripts/init_pack.py install …`).

**Do not commit** secrets, customer PII, or production KB dumps. Demo files only.

Legacy flat `data/{domain_id}/*.txt` (default tenant) is still **read** by discovery for older deploys; **new** content must use `data/{tenant}/{domain}/`.

After adding or changing files:

```bash
python scripts/reindex_rag.py
```
