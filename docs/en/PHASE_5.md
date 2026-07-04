# Phase 5 — Standard publication

**Goal:** Publish **Grounded as a checkable standard** — spec v1, conformance CLI, public benchmark, RFC governance.

**Branch:** `feature/phase-5-standard-publication`  
**Horizon:** 1 (Reference implementation)  
**Prerequisite:** Phase 4 merged to `main` ✅

---

## Pillars addressed

| Pillar | Phase 5 deliverables |
|--------|------------------------|
| **1 Spec & conformance** | [GROUNDED_SPEC_v1.md](./spec/GROUNDED_SPEC_v1.md), `python -m conformance check` |
| **2 Quality science** | [BENCHMARK.md](./BENCHMARK.md), `scripts/bench_report.py`, CI badge data |
| **5 Governance** | [RFC.md](./RFC.md), [RFC-0001](./rfcs/RFC-0001-grounded-compatible.md) |
| **3 Reference deploy** | Docs only (Terraform/adapters → Phase 6) |
| **4 Templates** | Docs only (legal pack → Phase 6) |

---

## Deliverables

| # | Item | Artifact | Done in Phase 5 |
|---|------|----------|-----------------|
| 1 | Normative spec v1 | `docs/en/spec/GROUNDED_SPEC_v1.md` | ✅ |
| 2 | Conformance CLI | `conformance/__main__.py` | ✅ |
| 3 | RFC process | `docs/en/RFC.md` | ✅ |
| 4 | RFC-0001 Grounded-compatible | `docs/en/rfcs/RFC-0001-grounded-compatible.md` | ✅ |
| 5 | Standard strategy doc | `STANDARD_STRATEGY.md` | ✅ |
| 6 | Benchmark report script | `scripts/bench_report.py` | ✅ |
| 7 | Release tag v0.3.0 | GitHub Release (manual) | pending |
| 8 | Public site grounded.dev | GitHub Pages | Phase 5b / 6 |

---

## Acceptance criteria

### Conformance CLI
```bash
pip install -r conformance/requirements.txt
python -m conformance spec          # offline OpenAPI
python -m conformance http --url http://127.0.0.1:8080
python -m conformance retrieval --rag-url http://127.0.0.1:5000/rag/context  # optional
python -m conformance check --url http://127.0.0.1:8080   # spec + http
```

### Benchmark
```bash
python scripts/bench_report.py --suite all
# writes eval/results/latest_bench.json summary
```

### RFC
- RFC-0001 status: **Accepted** (defines minimum «Grounded-compatible»)
- Future API changes use RFC template in `docs/en/rfcs/`

---

## Out of scope (Phase 6+)

- Reranker / hybrid search
- Vector DB adapters (Qdrant, pgvector)
- Terraform modules
- Legal FAQ template pack
- Hosted SaaS / billing

---

## Related

- [STANDARD_STRATEGY.md](./STANDARD_STRATEGY.md)
- [PHASE_4.md](./PHASE_4.md)
- [conformance/README.md](../../conformance/README.md)
