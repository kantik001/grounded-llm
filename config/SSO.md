# SSO (OIDC) — Admin panel

Phase B enterprise login for **admin UI** via OpenID Connect. Telegram and API keys for chat are unchanged.

SAML is not implemented in v1; use an OIDC bridge (Azure AD, Okta, Google Workspace, Keycloak).

## Enable

```bash
OIDC_ENABLED=true
OIDC_ISSUER=https://login.microsoftonline.com/{tenant-id}/v2.0
OIDC_CLIENT_ID=your-client-id
OIDC_CLIENT_SECRET=your-client-secret
OIDC_REDIRECT_URL=https://your-host/api/admin/auth/callback
OIDC_SESSION_SECRET=long-random-string   # or reuse ADMIN_SECRET
OIDC_ROLE_MAPPING_FILE=config/oidc_role_mapping.json
OIDC_SESSION_TTL_HOURS=12
OIDC_SCOPES=openid profile email
```

Register the redirect URL in your IdP app registration.

## Flow

1. User opens `/admin.html` → **Sign in with SSO**
2. `GET /api/admin/auth/login` → IdP
3. `GET /api/admin/auth/callback` → signed HttpOnly session cookie
4. Admin API accepts cookie (or legacy Basic Auth if enabled)

## Role mapping

Example: [oidc_role_mapping.json.example](./oidc_role_mapping.json.example)

- `default_roles` — if no group/email match (default: `kb_editor`)
- `groups` — IdP group name → RBAC roles (`admin`, `kb_editor`, `api_manager`, `chat_only`)
- `emails` — explicit email → roles
- `claim` — JWT claim for groups (default `groups`; Azure may use `roles`)

See [RBAC.md](./RBAC.md) for role permissions.

## Endpoints

| Method | Path | Auth |
|--------|------|------|
| GET | `/api/admin/auth/config` | Public |
| GET | `/api/admin/auth/login` | Public → redirect |
| GET | `/api/admin/auth/callback` | Public (IdP) |
| POST | `/api/admin/auth/logout` | Session cookie |

## Coexistence with Basic Auth

`ADMIN_PASSWORD` / `ADMIN_USERS_FILE` can stay enabled alongside OIDC. The admin UI shows SSO when `OIDC_ENABLED=true` and basic login when configured.

## Audit

SSO login/logout events are written to `audit_log` (`admin_login`, `admin_login_failed`, `admin_logout`).

## Local dev without IdP

Leave `OIDC_ENABLED=false` and use `ADMIN_PASSWORD` as before.
