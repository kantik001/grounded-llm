# Demo knowledge-base files (git only)

This tree holds **sample documents** shipped with the repo for eval and local demos.  
**Runtime source of truth** is Postgres (`kb_documents`) + blob store — not this directory.

## Layout (samples)

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

Official templates live under [`packs/*/data/`](../packs/). Install registers them into the registry:

```bash
python scripts/init_pack.py install default
curl -u admin:pass -X POST "http://localhost:8080/api/admin/ingest?domain_id=default" \
  -H "Content-Type: application/json" -d '{"sync": true}'
```

**Do not commit** secrets, customer PII, or production KB dumps.

## One-time migration

If you have files only under `data/` and an empty registry:

```bash
python scripts/backfill_kb_registry.py
```

Then run ingest. See [KB_SOURCE_OF_TRUTH.md](../docs/en/KB_SOURCE_OF_TRUTH.md).
