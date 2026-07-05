# Template packs

Official **template packs** ship as self-contained folders with a `pack.yaml` manifest.

## Layout

```text
packs/
  hr/
    pack.yaml       # manifest (v1)
    data/           # knowledge-base source files
    eval.jsonl      # retrieval eval baseline
  it_support/
    pack.yaml
    data/
    eval.jsonl
```

## pack.yaml (v1)

| Field | Description |
|-------|-------------|
| `pack` | Pack folder name (slug) |
| `version` | Semver string |
| `description` | Human-readable summary |
| `domain.id` | Domain id in `config/domains.json` |
| `domain.emoji`, `names`, `rag_k` | Domain catalog metadata |
| `locale` | Primary locale (`en` / `ru`) for prompts |
| `prompts` | `rag_system`, `rag_task_intro` |
| `onboarding` | Sample questions for web UI |
| `few_shot.general` | Retrieval few-shot example |
| `eval.suite` | Eval suite name → `eval/rag_{suite}_baseline.jsonl` |

## CLI

```bash
pip install pyyaml   # or: pip install -r api/requirements.txt

python scripts/init_pack.py list
python scripts/init_pack.py registry --validate
python scripts/init_pack.py registry --json
python scripts/init_pack.py install it_support
python scripts/init_pack.py install hr --tenant default

python scripts/init_pack.py new legal_faq --from hr
python scripts/init_pack.py new my_pack --install
```

`install` merges:

- `config/domains.json`
- `config/locales/{locale}/` (prompts, onboarding, few_shot)
- `data/{tenant}/{domain}/` (HR `default` uses legacy flat `data/default/`)
- `eval/rag_{suite}_baseline.jsonl`

Then:

```bash
python scripts/reindex_rag.py
python scripts/run_rag_eval.py --suite it_support
```

## Official packs

| Pack | Domain | Eval suite |
|------|--------|------------|
| `hr` | `default` | `default_en` |
| `it_support` | `it_support` | `it_support` |
| `legal_faq` | `legal_faq` | `legal_faq` |

Official registry: [registry.yaml](./registry.yaml) — validated in CI.

See [docs/en/domain-packs/](../docs/en/domain-packs/) for deploy guides.

**Contribute a pack:** [domain-pack-template/](../domain-pack-template/) · starter tasks in [GOOD_FIRST_ISSUES.md](../GOOD_FIRST_ISSUES.md).
