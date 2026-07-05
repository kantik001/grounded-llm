# Partner certification (outline)

**Status:** Draft — Phase 7 documentation. Program activates after public launch and first external integrators.

Goal: integrators demonstrate **Grounded-compatible** deployments without forking core.

---

## Levels (proposed)

| Level | Name | Requirements |
|-------|------|--------------|
| **L1** | Compatible deploy | Pass `python -m conformance check --url <prod>` |
| **L2** | Quality certified | L1 + retrieval eval ≥90% on official suites + adversarial gate |
| **L3** | Partner | L2 + 2+ customer refs, security questionnaire pack, support SLA |

---

## L1 checklist

```bash
pip install -r conformance/requirements.txt
python -m conformance spec
python -m conformance check --url https://customer.example.com
```

- HTTPS, `/ready` green, citations + verify on sample `/message`
- Document OpenAPI version and product tag in deployment runbook

---

## L2 checklist

- Run official eval suites against production RAG URL (read-only)
- Submit anonymized `eval/results/*.json` summary
- No regression vs published [BENCHMARK.md](./BENCHMARK.md) baselines

---

## L3 checklist

- Signed partner agreement (template TBD)
- Listed on partner page (post-launch)
- Annual re-certification on major releases

---

## Application process (future)

1. Open issue using label `partner-certification`
2. Attach conformance JSON + eval summary
3. Maintainers review within 10 business days
4. Badge: «Grounded-compatible L1/L2» in README / site

---

## Related

- [RFC-0001 Grounded-compatible](./rfcs/RFC-0001-grounded-compatible.md)
- [GOVERNANCE.md](./GOVERNANCE.md)
- [conformance/README.md](../../conformance/README.md)
