# `scripts/` — ops, CI helpers, pack CLI

Repo-root utilities. Prefer **Makefile** targets when they exist (`make help`). Deeper notes: [scripts-overview.md](../docs/en/knowledge-base/scripts-overview.md).

## RAG / eval

| Script | Purpose |
|--------|---------|
| `reindex_rag.py` | Rebuild vector (+ sparse) indexes from `data/` |
| `run_rag_eval.py` | Retrieval eval against `eval/*.jsonl` → `eval/results/` |
| `run_adversarial_e2e.py` | Adversarial E2E via Go `POST /message` (mocks) |
| `ci_eval_retrieval.sh` | CI/local gate: reindex + Python RAG + all retrieval suites |
| `bench_report.py` | Benchmark summary JSON from eval results |

```bash
python scripts/reindex_rag.py
make eval-retrieval SUITE=default_en
make eval-retrieval-ci
```

## Packs / domains

| Script | Purpose |
|--------|---------|
| `init_pack.py` | List / install / scaffold template packs |
| `pack_installer.py` | Installer library used by `init_pack` |
| `pack_registry.py` | Load + validate `packs/*/pack.yaml` |
| `init_domain.sh` / `init_domain.ps1` | Scaffold one domain + locale stubs + `data/` |

```bash
make init-pack-list
make init-pack-install PACK=it_support
```

See [`packs/README.md`](../packs/README.md).

## Smoke / load / CI server

| Script | Purpose |
|--------|---------|
| `smoke.sh` / `smoke.ps1` | Go API smoke (`make smoke`) |
| `load_smoke.sh` | Concurrent curl load smoke (`make load-smoke`) |
| `load_smoke.js` | k6 variant of load smoke |
| `llm_e2e_smoke.sh` | Real LLM + mocked RAG smoke |
| `backup_postgres_smoke.sh` | `pg_dump` / restore round-trip (`make backup-smoke`) |
| `ci_start_mock_server.sh` | Build/start Go server with LLM/RAG mocks for CI |

## Codegen / docs / site

| Script | Purpose |
|--------|---------|
| `gen_retriever_grpc.py` | Generate `api/gen/` from `api/proto/retriever.proto` (`--check` in CI) |
| `gen_architecture_png.py` | Optional text-only architecture PNG fallback |
| `generate_demo_gif.py` | README demo GIF |
| `build_site_data.py` | Static pack-registry JSON for `site/` |

## Other

| Script | Purpose |
|--------|---------|
| `sync_connector.py` | Pull docs into `data/` via a connector |
| `create_github_repo.ps1` | One-shot: create private GH repo + push |

## Makefile shortcuts

`make reindex` · `make eval-retrieval` · `make eval-retrieval-ci` · `make smoke` · `make load-smoke` · `make backup-smoke` · `make init-pack-list` · `make init-pack-install PACK=…`

Compose project name: **`grounded_llm`**.
