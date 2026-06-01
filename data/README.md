# Domain knowledge base

Place documents per domain (plain text, PDF, or Word):

```
data/
  {domain_id}/
    document1.txt
    policy.pdf
    handbook.docx
```

Supported formats: `.txt`, `.pdf`, `.docx` (UTF-8 for text files).

After adding or changing files, reindex: `python scripts/reindex_rag.py`
