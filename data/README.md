# Domain knowledge base

Place documents per tenant and domain (plain text, PDF, or Word):

```
data/
  {tenant_id}/
    {domain_id}/
      document1.txt
      policy.pdf
```

Legacy layout `data/{domain_id}/` is still supported for the `default` tenant.

**Official template packs:**

| Domain ID | Path | Doc |
|-----------|------|-----|
| `default` (HR demo) | `data/default/*_en.txt` | [HR pack](../docs/en/domain-packs/HR.md) |
| `it_support` | `data/default/it_support/` | [IT Support pack](../docs/en/domain-packs/IT_SUPPORT.md) |

Supported formats: `.txt`, `.pdf`, `.docx` (UTF-8 for text files).

After adding or changing files, reindex: `python scripts/reindex_rag.py`
