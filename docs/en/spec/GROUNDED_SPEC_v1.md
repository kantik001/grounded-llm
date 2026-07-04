# Grounded Spec v1.0

**Status:** Normative for API v1  
**OpenAPI:** `server/openapi.v1.json` (live: `GET /api/v1/openapi.json`)  
**Policy:** [API_DEPRECATION_POLICY.md](../API_DEPRECATION_POLICY.md)  
**Conformance:** [conformance/README.md](../../conformance/README.md)

---

## 1. Scope

Grounded Spec v1 defines a **document-grounded chat API** where:

1. Answers are produced from **retrieved document chunks** (RAG).
2. Responses include **citations** (source filename + excerpt).
3. **Numeric claims** in answers are verified against retrieved context (reference implementation).
4. **Multi-tenant** isolation is supported via `X-Tenant-ID`.

Out of scope for v1: arbitrary tool use, code execution, visual workflow graphs.

---

## 2. Terminology

| Term | Definition |
|------|------------|
| **Platform** | Go orchestration service (default port 8080) |
| **RAG service** | Python retrieval service (default port 5000) |
| **Domain** | Knowledge workspace (`domain_id`, e.g. `default`, `it_support`) |
| **Tenant** | Isolation boundary (`tenant_id`, default `default`) |
| **Template pack** | Config + documents + eval under `data/{tenant}/{domain}/` |

---

## 3. Authentication (normative)

Integrators MUST support at least one of:

| Method | Header | Use case |
|--------|--------|----------|
| API key | `X-API-Key` | Programmatic access |
| Telegram Web App | `X-Telegram-Init-Data` | Mini App (optional) |

Protected routes MUST return `401` when auth is missing or invalid.

Admin routes (`/api/admin/*`) MUST require admin credentials (Basic and/or OIDC session).

---

## 4. Public endpoints (no auth)

Implementations MUST expose:

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/health` | Liveness |
| GET | `/ready` | Readiness (DB + RAG when not mocked) |
| GET | `/api/v1/openapi.json` | Machine-readable contract |
| GET | `/api/domains` | List domains |
| GET | `/api/branding` | UI branding |
| GET | `/api/onboarding` | Suggested questions |

Unversioned aliases (`/session`, `/message`) MAY exist but `/api/v1/*` is the stable integrator surface.

---

## 5. Chat flow (normative)

### 5.1 Create session

```http
POST /api/v1/session
Content-Type: application/json
X-API-Key: <key>
X-Tenant-ID: default

{"domain_id": "default"}
```

Response MUST include `session_id`.

### 5.2 Send message

```http
POST /api/v1/message
Content-Type: application/json

{"session_id": "...", "domain_id": "default", "text": "..."}
```

Response MUST include `success`, updated `messages[]`. Assistant messages SHOULD include `citations[]` with at least `filename` when context was retrieved.

Streaming (`?stream=1`) MAY emit SSE events; event shapes are part of v1 stability per deprecation policy.

---

## 6. Verification (reference behavior)

When retrieval returns fragments, the reference platform:

1. Generates an answer from LLM using retrieved context.
2. Runs **numeric verify**: numbers in the answer must appear in fragment text.
3. On verify failure, returns a user-visible warning and still MAY include citations.

Alternate implementations claiming **Grounded-compatible** MUST document verify behavior; see [RFC-0001](../rfcs/RFC-0001-grounded-compatible.md).

---

## 7. Retrieval (internal contract)

Platform calls Python:

```http
POST /rag/context
{"question", "domain_id", "tenant_id", "locale"}
```

Production deployments SHOULD protect this with internal token (`X-RAG-Service-Token`). This path is **not** part of public integrator v1.

---

## 8. Conformance levels

| Level | Requirements |
|-------|----------------|
| **Core API** | Public + session + message paths match OpenAPI; conformance `spec` + `http` pass |
| **Grounded-compatible** | Core + citations on grounded answers + documented verify + golden retrieval eval pass |
| **Reference impl** | This repository + all CI gates green |

Run checks:

```bash
python -m conformance check --url https://your-host:8080
```

---

## 9. Compatibility

See [COMPATIBILITY.md](../COMPATIBILITY.md) for supported Go/Python/Postgres versions and embedding model pin.

---

## 10. Changes

Spec changes follow [RFC.md](../RFC.md). Breaking API changes require `/api/v2` per [API_DEPRECATION_POLICY.md](../API_DEPRECATION_POLICY.md).

---

## Related

- [GROUNDED_SPEC_v1.md](./GROUNDED_SPEC_v1.md) — this document
- OpenAPI: `/api/v1/openapi.json`
- [RFC-0001](../rfcs/RFC-0001-grounded-compatible.md)
