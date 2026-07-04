# RFC process

Grounded LLM uses lightweight **RFCs** (Request for Comments) for spec changes, conformance levels, and cross-cutting architecture decisions.

---

## When to write an RFC

| Need RFC | Examples |
|----------|----------|
| Yes | New API v1 fields that affect clients; conformance level changes; new normative behavior |
| Yes | Breaking changes (must also plan `/api/v2`) |
| No | Bug fixes matching existing spec |
| No | New template pack or eval cases |
| No | Internal refactors without API change |

---

## Lifecycle

```text
Draft → Review (PR + discussion) → Accepted → Implemented → Superseded
```

| Status | Meaning |
|--------|---------|
| **Draft** | Work in progress, not normative |
| **Review** | PR open, seeking feedback |
| **Accepted** | Normative; implementation may proceed |
| **Implemented** | Merged to `main` with tests/docs |
| **Superseded** | Replaced by newer RFC |

---

## How to submit

1. Copy `docs/en/rfcs/RFC-0000-template.md` → `RFC-NNNN-short-title.md`
2. Fill sections: Summary, Motivation, Specification, Conformance impact, Alternatives
3. Open PR with label `rfc` (or note in PR body)
4. Minimum **7 days** review for Accepted (or maintainer fast-track for typos/clarifications)
5. On merge: update [GROUNDED_SPEC_v1.md](./spec/GROUNDED_SPEC_v1.md) or CHANGELOG if behavior changed

---

## Steering (initial)

Until a formal committee exists:

- **Maintainers** listed in README / GitHub org approve Accepted RFCs
- Disputes: GitHub Discussion → documented decision in RFC header

Future: publish `STEERING.md` with named maintainers and election process (Horizon 3).

---

## Index

| RFC | Title | Status |
|-----|-------|--------|
| [RFC-0001](./rfcs/RFC-0001-grounded-compatible.md) | Grounded-compatible definition | Accepted |

---

## Related

- [STANDARD_STRATEGY.md](./STANDARD_STRATEGY.md)
- [API_DEPRECATION_POLICY.md](./API_DEPRECATION_POLICY.md)
- [PHASE_5.md](./PHASE_5.md)
