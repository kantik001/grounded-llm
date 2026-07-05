# Phase 7 — Platform ecosystem

**Goal:** Pack registry, retrieval quality (cross-encoder), ingest connectors, GCP deploy, governance.

**Branch:** `feature/phase-7-platform-ecosystem` — **merged to `main`**  
**Horizon:** 2 (Platform standard)  
**Prerequisite:** Phase 6 merged to `main` ✅

---

## Pillars addressed

| Pillar | Phase 7 deliverables |
|--------|------------------------|
| **4 Template marketplace** | `packs/registry.yaml` + validate CLI |
| **2 Quality science** | Cross-encoder reranker (`RAG_RERANKER=cross_encoder`) |
| **3 Reference deploy** | GCP Terraform (`deploy/terraform/gcp/reference/`) |
| **4 + ecosystem** | Connector interface + `local_folder` reference |
| **5 Governance** | [GOVERNANCE.md](./GOVERNANCE.md), [PARTNER_CERTIFICATION.md](./PARTNER_CERTIFICATION.md) |

---

## Deliverables

| # | Item | Artifact |
|---|------|----------|
| 1 | Pack registry | `packs/registry.yaml`, `init_pack.py registry` |
| 2 | Cross-encoder rerank | `rag/rerank.py`, `RAG_RERANKER` |
| 3 | Ingest connectors | `connectors/`, [CONNECTORS.md](./CONNECTORS.md) |
| 4 | GCP Terraform | `deploy/terraform/gcp/reference/` |
| 5 | Governance | GOVERNANCE + partner certification outline |

---

## Acceptance criteria

### Registry
```bash
python scripts/init_pack.py registry --validate
python scripts/init_pack.py registry --json
```

### Cross-encoder (optional, local)
```bash
RAG_RERANKER=cross_encoder python scripts/run_rag_eval.py --suite legal_faq
```

### Connector
```bash
python scripts/sync_connector.py local_folder --source ./packs/hr/data --domain default --dry-run
```

### GCP Terraform
```bash
cd deploy/terraform/gcp/reference
terraform init && terraform validate
```

---

## Out of scope (Phase 8+)

- Hosted SaaS / billing
- Pack registry web UI
- SharePoint / Drive / Confluence connectors (implementation)
- Azure Terraform
- Public repo + GitHub Pages launch

---

## Related

- [PHASE_6.md](./PHASE_6.md)
- [STANDARD_STRATEGY.md](./STANDARD_STRATEGY.md)
