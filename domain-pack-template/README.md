# Domain pack template

Copy this layout for a new customer / industry pack:

```
config/
  domains.json      # add domain entry
  prompts.json
  few_shot.json
  onboarding.json
  branding.json
data/
  {tenant_id}/
    {domain_id}/
      *.txt | *.pdf | *.docx
eval/
  rag_{domain_id}_baseline.jsonl
```

Quick start:

```bash
./scripts/init_domain.sh hr_handbook default
# edit config/*.json, add documents, then:
python scripts/reindex_rag.py
```
