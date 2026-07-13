# Phase 9 — Launch & live connectors

**Goal:** Live Google Drive + Confluence connectors, hosted/billing prep, public launch playbook.

**Branch:** `feature/phase-9-launch-scale` — **merged to `main`**  
**Horizon:** 2 → 3 (launch prep)  
**Prerequisite:** Phase 8 merged to `main` ✅

---

## Pillars addressed

| Pillar | Phase 9 deliverables |
|--------|------------------------|
| **4 Ecosystem** | Google Drive + Confluence REST connectors |
| **3 Platform** | Plan tiers scaffold (`config/plans.yaml`) |
| **Adoption** | [LAUNCH.md](./LAUNCH.md) public release playbook |
| **Path B prep** | [SAAS.md](./SAAS.md), [BILLING.md](./BILLING.md) |

---

## Deliverables

| # | Item | Artifact |
|---|------|----------|
| 1 | Google Drive API | `connectors/google_drive.py` |
| 2 | Confluence REST | `connectors/confluence.py` |
| 3 | Connector deps | `api/requirements-connectors.txt` |
| 4 | Billing scaffold | `config/plans.yaml`, BILLING.md |
| 5 | Launch playbook | LAUNCH.md |
| 6 | Site packs UI | `site/index.html` + `packs.json` |

---

## Acceptance criteria

### Confluence (dry-run with credentials)
```bash
python scripts/sync_connector.py confluence --domain it_support --dry-run
```

### Google Drive
```bash
pip install -r api/requirements-connectors.txt
python scripts/sync_connector.py google_drive --domain default --dry-run
```

### Plans scaffold
```bash
python -c "import yaml; yaml.safe_load(open('config/plans.yaml'))"
```

---

## Out of scope (Phase 10+)

- Stripe webhook implementation
- Self-serve signup UI
- Making repository public (operator decision — see LAUNCH.md)

---

## Related

- [PHASE_8.md](./PHASE_8.md)
- [LAUNCH.md](./LAUNCH.md)
