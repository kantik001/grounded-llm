# Phase 8 — Connectors & multi-cloud

**Goal:** Enterprise ingest connectors, Azure deploy reference, intranet embed widget.

**Branch:** `feature/phase-8-connectors-cloud` — **merged to `main`**  
**Horizon:** 2 (Platform standard)  
**Prerequisite:** Phase 7 merged to `main` ✅

---

## Pillars addressed

| Pillar | Phase 8 deliverables |
|--------|------------------------|
| **4 Template marketplace** | `site/packs.json` registry export |
| **3 Reference deploy** | Azure Terraform (`deploy/terraform/azure/reference/`) |
| **4 + ecosystem** | SharePoint Graph + export connectors |
| **Adoption** | Embeddable chat widget (`webapp/embed.html`) |

---

## Deliverables

| # | Item | Artifact |
|---|------|----------|
| 1 | Export connectors | `sharepoint_export`, `google_drive_export`, `confluence_export` |
| 2 | SharePoint Graph | `connectors/sharepoint.py` + env vars |
| 3 | Azure Terraform | `deploy/terraform/azure/reference/` |
| 4 | Embed widget | [EMBED.md](./EMBED.md), `webapp/embed.*` |
| 5 | Site data | `site/packs.json`, `scripts/build_site_data.py` |
| 6 | Pages CI | workflow manual-only until repo is public |

---

## Acceptance criteria

### Export connector
```bash
python scripts/sync_connector.py sharepoint_export --source ./packs/hr/data --domain default --dry-run
```

### SharePoint Graph (requires Azure AD app + drive id)
```bash
export SHAREPOINT_TENANT_ID=...
export SHAREPOINT_CLIENT_ID=...
export SHAREPOINT_CLIENT_SECRET=...
export SHAREPOINT_DRIVE_ID=...
python scripts/sync_connector.py sharepoint --domain it_support --dry-run
```

### Azure Terraform
```bash
cd deploy/terraform/azure/reference
terraform init && terraform validate
```

### Embed
Open `http://localhost/embed.html?api=/api/` after `docker compose up`.

---

## Out of scope (Phase 9+)

- Hosted SaaS / billing
- Google Drive live API connector
- Confluence live API connector
- Public repo launch + PR campaign

---

## Related

- [PHASE_7.md](./PHASE_7.md)
- [CONNECTORS.md](./CONNECTORS.md)
