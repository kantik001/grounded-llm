# Phase 6 — Ecosystem scale

**Goal:** Extend the reference platform with **template growth**, **retrieval options**, and **cloud deploy** primitives.

**Branch:** `feature/phase-6-ecosystem-scale`  
**Horizon:** 2 (Platform standard)  
**Prerequisite:** Phase 5 merged to `main` ✅

---

## Pillars addressed

| Pillar | Phase 6 deliverables |
|--------|------------------------|
| **3 Reference deploy** | AWS Terraform reference (`deploy/terraform/aws/reference/`) |
| **4 Template marketplace** | Legal FAQ official pack |
| **2 Quality science** | Hybrid keyword reranking (`RAG_RETRIEVAL_MODE=hybrid`) |
| **1 Spec & conformance** | Vector store adapter (`VECTOR_STORE=chroma\|qdrant`) |

---

## Deliverables

| # | Item | Artifact |
|---|------|----------|
| 1 | Legal FAQ template pack | `packs/legal_faq/` |
| 2 | Vector store adapter | `rag/vector_backend/` |
| 3 | Qdrant optional deps | `api/requirements-qdrant.txt` |
| 4 | Hybrid retrieval | `rag/hybrid_rank.py`, `RAG_RETRIEVAL_MODE` |
| 5 | AWS Terraform reference | `deploy/terraform/aws/reference/` |
| 6 | Docs | [LEGAL_FAQ.md](./domain-packs/LEGAL_FAQ.md), [VECTOR_STORE.md](./VECTOR_STORE.md), [TERRAFORM.md](./TERRAFORM.md) |

---

## Acceptance criteria

### Legal FAQ pack
```bash
python scripts/init_pack.py install legal_faq
python scripts/reindex_rag.py
python scripts/run_rag_eval.py --suite legal_faq
```

### Hybrid retrieval
```bash
RAG_RETRIEVAL_MODE=hybrid python scripts/run_rag_eval.py --suite legal_faq
```

### Vector adapter
```bash
# default (CI)
VECTOR_STORE=chroma pytest tests/test_vector_backend.py -v

# optional Qdrant
pip install -r api/requirements-qdrant.txt
VECTOR_STORE=qdrant QDRANT_URL=http://127.0.0.1:6333 python scripts/reindex_rag.py
```

### Terraform
```bash
cd deploy/terraform/aws/reference
terraform init
terraform validate
```

---

## Out of scope (Phase 7+)

- Hosted SaaS / billing
- Pack public registry UI
- Cross-encoder reranker model
- GCP/Azure Terraform modules

---

## Related

- [STANDARD_STRATEGY.md](./STANDARD_STRATEGY.md)
- [PHASE_5.md](./PHASE_5.md)
- [packs/README.md](../../packs/README.md)
