# Template pack scaffold

Use this layout to add a new **document-grounded assistant** on Grounded LLM.

A template pack = config + documents + eval + locale strings. The platform core (`server/`, `api/`, `rag/`) stays unchanged.

## Layout

```
config/
  domains.json              # add domain entry
  locales/{en}/               # prompts, branding, onboarding, few_shot
data/
  {tenant_id}/
    {domain_id}/
      *.txt | *.pdf | *.docx
eval/
  rag_{domain_id}_baseline.jsonl
```

## Quick start

```bash
./scripts/init_domain.sh hr_policies default
# 1. Add domain entry to config/domains.json
# 2. Copy locale files from config/locales/en/ and customize
# 3. Add documents to data/default/hr_policies/
# 4. Add eval cases to eval/rag_hr_policies_baseline.jsonl
python scripts/reindex_rag.py
python scripts/run_rag_eval.py --suite default_en   # or your suite
```

## Reference template

Full example: [docs/en/domain-packs/HR.md](../docs/en/domain-packs/HR.md)

## Docs

- [PLATFORM_VISION.md](../PLATFORM_VISION.md)
- [docs/en/ARCHITECTURE.md](../docs/en/ARCHITECTURE.md)
- [docs/en/LOCALE_GUIDE.md](../docs/en/LOCALE_GUIDE.md)
