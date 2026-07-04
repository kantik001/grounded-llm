# Phase 4 — Spec, trust & conformance

**Goal:** Move from «enterprise-ready deploy» to **contract-stable platform** — API policy, reproducible conformance, adversarial quality gates, and GDPR-style tenant purge.

**Status:** ✅ Complete (merged to `main`)

**Branch:** `feature/phase-4-spec-trust`

---

## Deliverables

| # | Item | Artifact | CI job (target) |
|---|------|----------|-----------------|
| 1 | API deprecation policy | [API_DEPRECATION_POLICY.md](./API_DEPRECATION_POLICY.md) | docs review |
| 2 | Compatibility matrix | [COMPATIBILITY.md](./COMPATIBILITY.md) | pinned in Docker/CI |
| 3 | Conformance test suite | [conformance/](../../conformance/) | `conformance-spec` + HTTP in `smoke-api` |
| 4 | Secret scanning | [.github/workflows/secret-scan.yml](../../.github/workflows/secret-scan.yml) | `secret-scan` |
| 5 | Adversarial eval pack | [eval/rag_adversarial_baseline.jsonl](../../eval/rag_adversarial_baseline.jsonl) | `eval-retrieval-gate` |
| 5b | Adversarial E2E | [eval/rag_adversarial_e2e.jsonl](../../eval/rag_adversarial_e2e.jsonl) | `smoke-api` |
| 6 | Tenant data purge | `DELETE /api/admin/tenants/:id` | `go-test` + audit |

---

## Implementation order (recommended)

```text
Week 1   Secret scan (CI) + compatibility matrix in README/CI pins
Week 1   API deprecation policy published + OpenAPI info.version alignment
Week 2   Conformance HTTP tests (public + v1 paths) against running server
Week 2   Adversarial JSONL validation + retrieval checks (expect_not_contains)
Week 3   Adversarial E2E (/message verify + citations) — nightly or staging
Week 3–4 Tenant purge admin endpoint + audit + Trust Center update
```

---

## Acceptance criteria

### 1. API deprecation policy
- Document published at `docs/en/API_DEPRECATION_POLICY.md`
- OpenAPI `info.version` matches policy (`1.x`)
- CHANGELOG references policy for any breaking change

### 2. Compatibility matrix
- Single source: `docs/en/COMPATIBILITY.md`
- CI uses Go 1.23, Python 3.11, Postgres 16 (already true)
- Embedding model pin documented: `intfloat/multilingual-e5-small`

### 3. Conformance suite
- Runnable without proprietary keys: `LLM_MOCK=true RAG_MOCK=true`
- Validates OpenAPI paths vs live HTTP status codes
- Golden retrieval: wraps `scripts/run_rag_eval.py --suite all`
- README for third-party implementors in `conformance/README.md`

### 4. Secret scanning
- gitleaks on every PR + push to `main`
- No secrets in repo history (baseline scan on first run)
- CONTRIBUTING mentions secret scan failure workflow

### 5. Adversarial eval pack
- ≥20 cases in `eval/rag_adversarial_baseline.jsonl`
- Categories: wrong_number, missing_citation, cross_domain, prompt_injection, out_of_scope
- Phase 4 extends `run_rag_eval.py` for `expect_not_contains` + optional E2E runner for verify/citations

### 6. Tenant purge (RTBF)
- Admin-only `DELETE /api/admin/tenants/:tenant_id?confirm=true`
- Deletes: Postgres rows (sessions, messages, feedback, audit for tenant), `data/{tenant}/`, upload refs
- Audit log entry before delete; documented in Trust Center
- See [TENANT_PURGE.md](./TENANT_PURGE.md) for API contract (implement in Phase 4)

---

## Out of scope (Phase 4)

- Vector DB adapters (Phase C)
- Full SOC2 / pentest
- RFC steering committee (lightweight policy doc only)
- Conformance certification trademark

---

## Related

- [ROADMAP.md](./ROADMAP.md)
- [TRUST_CENTER.md](./TRUST_CENTER.md)
- [API_DEPRECATION_POLICY.md](./API_DEPRECATION_POLICY.md)
- [COMPATIBILITY.md](./COMPATIBILITY.md)
