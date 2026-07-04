# RFC-0001 — Grounded-compatible

**RFC:** 0001  
**Title:** Definition of «Grounded-compatible»  
**Status:** Accepted  
**Authors:** Grounded LLM maintainers  
**Created:** 2026-07-05  
**Pillar:** 1 (Spec & conformance), 2 (Quality science)

---

## Summary

This RFC defines the minimum requirements for an implementation to claim **Grounded-compatible** with Spec v1 — beyond passing HTTP/OpenAPI checks.

---

## Motivation

Integrators and vendors need a **testable label**, not marketing. «Uses RAG» is meaningless. Grounded-compatible means citations + measurable retrieval quality + documented verify behavior.

---

## Specification

An implementation is **Grounded-compatible** when all of the following hold:

### 1. Core API (MUST)

- Passes `python -m conformance spec`
- Passes `python -m conformance http --url <base>` against a running deployment
- Exposes `/api/v1/openapi.json` consistent with implemented behavior
- Supports `X-API-Key` (or documented equivalent) on protected routes

### 2. Citations (MUST)

- `POST /api/v1/message` responses include `citations[]` on assistant messages when retrieval returned fragments
- Each citation includes at least `filename` (or equivalent source id) and human-readable excerpt or content reference

### 3. Verify (MUST document)

- Implementation MUST document how numeric claims in answers are checked against sources
- Reference behavior: numbers in final answer must appear in retrieved fragment text; failures surface a visible warning (see [GROUNDED_SPEC_v1.md](../spec/GROUNDED_SPEC_v1.md) §6)
- Alternate algorithms allowed if documented and adversarial eval pass rate is published

### 4. Retrieval quality (MUST)

- Passes golden retrieval eval: `python -m conformance retrieval --rag-url <url>` with reference JSONL suites shipped in `eval/` (or vendor-published equivalent suites approved by RFC)
- Minimum pass rate: **100%** on `default_en`, `it_support`, and `adversarial` suites for a release tagged Grounded-compatible

### 5. Multi-tenant (SHOULD)

- Honors `X-Tenant-ID` for session/message isolation when multi-tenant mode is enabled

---

## Conformance levels (normative)

| Level | Label | Requirements |
|-------|-------|--------------|
| L1 | API core | §1 only |
| L2 | **Grounded-compatible** | §1–4 |
| L3 | Reference implementation | This repository, all CI jobs green |

Products MUST NOT use «Grounded-compatible» for L1-only deployments.

---

## Alternatives considered

| Alternative | Rejected because |
|-------------|------------------|
| Trademark-only certification | Not enforceable in OSS without legal entity |
| LLM-judge quality gate | Non-deterministic, expensive in CI |
| No verify requirement | Core differentiator lost |

---

## Implementation plan

- [x] Conformance CLI (`python -m conformance`)
- [x] Golden retrieval in conformance
- [ ] Public registry of compatible products (Phase 6)
- [ ] Certification exam for integrators (Phase 7)

---

## Related

- [GROUNDED_SPEC_v1.md](../spec/GROUNDED_SPEC_v1.md)
- [BENCHMARK.md](../BENCHMARK.md)
- [conformance/README.md](../../../conformance/README.md)
