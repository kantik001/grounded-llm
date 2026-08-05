# Config examples (templates)

Copy into `config/` (parent directory) and point env vars at the copies. Do not commit real secrets.

| Template | Typical dest | Env |
|----------|--------------|-----|
| `tenants.json.example` | `config/tenants.json` | `TENANTS_REGISTRY_FILE` |
| `tenant_quotas.json.example` | `config/tenant_quotas.json` | `TENANT_QUOTAS_FILE` |
| `admin_users.json.example` | `config/admin_users.json` | `ADMIN_USERS_FILE` |
| `api_keys.json.example` | `config/api_keys.json` | `API_KEYS_FILE` |
| `oidc_role_mapping.json.example` | `config/oidc_role_mapping.json` | `OIDC_ROLE_MAPPING_FILE` |

See [../README.md](../README.md).
