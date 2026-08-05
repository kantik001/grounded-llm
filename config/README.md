# `config/` — runtime configuration

Mounted into containers as `/config` (no image rebuild). Go (`server/`) and Python (`rag/`, `api/`) read from here.

Deep dive: [docs/en/knowledge-base/config-overview.md](../docs/en/knowledge-base/config-overview.md).

## Layout

| Path | Role |
|------|------|
| `domains.json` | Knowledge-domain catalog (`DOMAINS_CONFIG_PATH`) |
| `locales/{ru,en}/` | Prompts, branding, onboarding, few-shot (`LOCALES_ROOT`) — see [locales/README.md](./locales/README.md) |
| `plans.yaml` | SaaS plan tiers (`PLANS_FILE`) |
| `examples/` | Safe templates for optional operator files (copy → repo root of `config/`) |
| `schemas/` | JSON Schema for CI / local validation |
| `*.md` | Operator notes: SSO, RBAC, quotas, reindex, analytics |

## Environment

| Variable | Default / typical | Notes |
|----------|-------------------|--------|
| `DOMAINS_CONFIG_PATH` | `config/domains.json` | Domain catalog |
| `LOCALES_ROOT` | `config/locales` | Locale bundles |
| `DEFAULT_LOCALE` | `en` | Fallback when request has no locale |
| `PLANS_FILE` | `config/plans.yaml` | Billing / plan list |
| `TENANTS_REGISTRY_FILE` | — | e.g. `config/tenants.json` (from example) |
| `TENANT_QUOTAS_FILE` | — | e.g. `config/tenant_quotas.json` |
| `ADMIN_USERS_FILE` | — | e.g. `config/admin_users.json` |
| `API_KEYS_FILE` | — | e.g. `config/api_keys.json` |
| `OIDC_ROLE_MAPPING_FILE` | — | e.g. `config/oidc_role_mapping.json` |

Runtime secrets and live registries live as **non-`.example` files** (often gitignored locally). Templates stay under `examples/`.

```bash
cp config/examples/tenants.json.example config/tenants.json
cp config/examples/tenant_quotas.json.example config/tenant_quotas.json
```

## Validation

```bash
# Covered by pytest (CI python-test job)
pytest tests/test_domains_schema.py -v
```

Schema: [`schemas/domains.schema.json`](./schemas/domains.schema.json).

## Related docs

- [SSO.md](./SSO.md) · [RBAC.md](./RBAC.md) · [QUOTAS.md](./QUOTAS.md)
- [REINDEX.md](./REINDEX.md) · [ANALYTICS.md](./ANALYTICS.md)
