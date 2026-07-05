# Governance

How **Grounded LLM** is governed as an open platform standard — roles, decisions, and release expectations.

See also: [RFC.md](./RFC.md) · [CONTRIBUTING.md](../../CONTRIBUTING.md) · [STANDARD_STRATEGY.md](./STANDARD_STRATEGY.md)

---

## Roles

| Role | Responsibility |
|------|----------------|
| **Maintainers** | Merge PRs, cut releases, CI health, security response |
| **Contributors** | Issues, PRs, packs, docs, conformance fixes |
| **Integrators** | Deploy on customer infra; may fork for private patches |
| **Partners** | Certified deployments (see [PARTNER_CERTIFICATION.md](./PARTNER_CERTIFICATION.md)) |

Current maintainers: listed in [README.md](../../README.md) (update before public launch).

---

## Decision process

| Change type | Process |
|-------------|---------|
| Bug fix, docs, tests | PR + review |
| New template pack | PR + eval baseline + registry entry |
| API behavior change | [RFC](./RFC.md) required |
| Spec v1 breaking change | RFC + semver major + migration guide |
| Release tag `v*.*.*` | Green CI on `main`, CHANGELOG, [RELEASE.md](./RELEASE.md) |

RFC template: [rfcs/RFC-0000-template.md](./rfcs/RFC-0000-template.md)

---

## Release cadence (target)

| Track | Cadence | Contents |
|-------|---------|----------|
| **Patch** | as needed | fixes, dependency bumps |
| **Minor** | ~6–8 weeks | packs, adapters, non-breaking features |
| **Major** | rare | spec/API breaking, embedding model change |

Every release with retrieval changes must pass `eval-retrieval-gate` on `main`.

---

## Steering (Horizon 3)

When the project goes public and gains external integrators:

- Form a **steering group** (maintainers + 2–3 partner reps)
- Quarterly roadmap review aligned with [ROADMAP.md](./ROADMAP.md) pillars
- RFC acceptance by steering for spec-level changes

Not active until public launch — document now for transparency.

---

## Code of conduct & security

- [CODE_OF_CONDUCT.md](../../CODE_OF_CONDUCT.md)
- [SECURITY.md](../../SECURITY.md) — report vulnerabilities privately

---

## Related

- [PHASE_7.md](./PHASE_7.md)
- [API_DEPRECATION_POLICY.md](./API_DEPRECATION_POLICY.md)
