# Template pack scaffold

Use **`python scripts/init_pack.py`** to install or create template packs.

## Install an official pack

```bash
python scripts/init_pack.py list
python scripts/init_pack.py install it_support
python scripts/init_pack.py install hr
```

## Create a new pack

```bash
python scripts/init_pack.py new legal_faq --from hr
# edit packs/legal_faq/pack.yaml and packs/legal_faq/data/
python scripts/init_pack.py install legal_faq
python scripts/reindex_rag.py
python scripts/run_rag_eval.py --suite legal_faq
```

## Manual layout (advanced)

See [packs/README.md](../packs/README.md) for `pack.yaml` v1 spec.

Legacy helper (data dir only):

```bash
./scripts/init_domain.sh my_domain default
```
