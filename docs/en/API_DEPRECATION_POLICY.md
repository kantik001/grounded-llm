# API deprecation policy

Grounded LLM exposes a **versioned HTTP API** for integrators. This document defines stability guarantees, breaking changes, and sunset timelines.

**Current stable version:** `/api/v1/*`  
**OpenAPI:** `GET /api/v1/openapi.json`  
**Product semver:** repository tags (`v0.x.y`) — independent from API version number.

---

## Versioning model

| Surface | Path prefix | Stability |
|---------|-------------|-----------|
| **Stable integrator API** | `/api/v1/*` | Breaking changes require `/api/v2` |
| **Unversioned aliases** | `/api/*`, `/session`, `/message`, … | Mirror v1 behavior; may lag documentation |
| **Admin API** | `/api/admin/*` | Stable for self-hosted ops; breaking changes noted in CHANGELOG |
| **Internal** | Python `/rag/context` | Not a public contract; change with server releases |

We follow [Semantic Versioning](https://semver.org/) for **repository releases**. API path version (`v1`) increments only on **incompatible** HTTP contract changes.

---

## What counts as breaking

Breaking changes **require a new API version** (`/api/v2`):

- Removing or renaming an endpoint under `/api/v1`
- Changing required request fields or authentication scheme
- Changing response JSON shape in a way that breaks typed clients
- Changing HTTP status code semantics for the same error condition
- Removing enum values or citation/verify fields that clients depend on

**Non-breaking** (allowed in `/api/v1` with minor release notes):

- Adding optional request/response fields
- Adding new endpoints
- Adding new optional query parameters
- Performance improvements with identical responses
- Stricter validation that rejects previously invalid input

---

## Deprecation process

When we need to remove or replace behavior in `/api/v1`:

1. **Announce** — CHANGELOG `[Unreleased]` + GitHub Discussion or release notes
2. **Mark deprecated** — OpenAPI `deprecated: true` + `Sunset` response header on affected routes
3. **Migration doc** — update `docs/en/API_EXAMPLES.md` with replacement flow
4. **Minimum notice** — **6 months** after deprecation announcement before removal from `/api/v1`
5. **Remove** — only in `/api/v2` (new path) or major product release with migration guide

### Sunset header

Deprecated responses SHOULD include:

```http
Sunset: Sat, 01 Jan 2028 00:00:00 GMT
Deprecation: true
Link: </api/v2/openapi.json>; rel="successor-version"
```

---

## `/api/v1` stability commitment

While `/api/v1` is the current stable integrator API:

- We maintain **backwards-compatible** JSON for documented fields in OpenAPI
- SDK (`grounded-llm` Python package) targets `/api/v1` only
- CI **conformance suite** (`conformance/`) must pass before release tags

Legacy unversioned paths (`/message`, `/session`) remain for Web App and older clients but **new integrations MUST use `/api/v1`**.

---

## Introducing `/api/v2`

When breaking change is unavoidable:

1. Ship `/api/v2/*` alongside `/api/v1/*`
2. Publish `openapi.v2.json` and conformance tests for v2
3. Keep `/api/v1` for at least **12 months** after v2 GA (unless security-critical)
4. Default OpenAPI link in README points to latest stable (v2) after migration period starts

---

## Streaming (`?stream=1`)

SSE event shapes are part of the v1 contract. New event types may be added; existing `delta`, `done`, `error` events will not change meaning without a version bump.

---

## Reporting issues

- **Bug** (implementation differs from OpenAPI): GitHub issue with `bug` template
- **Spec gap** (undocumented behavior): feature request
- **Breaking change proposal**: open a discussion before PR; reference this policy

See [CONTRIBUTING.md](../../CONTRIBUTING.md) and [CHANGELOG.md](../../CHANGELOG.md).
