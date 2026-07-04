# Security Policy

## Supported versions

| Version | Supported |
|---------|-----------|
| `main` branch | ✅ Active development |
| Latest release tag | ✅ When published |
| Older releases | ⚠️ Best effort |

## Reporting a vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Use one of these channels:

1. **GitHub Security Advisories (preferred):** [Report a vulnerability](https://github.com/kantik001/grounded-llm/security/advisories/new) on this repository.
2. **Private contact:** Open a [private security advisory](https://github.com/kantik001/grounded-llm/security/advisories) or contact the maintainer via GitHub ([@kantik001](https://github.com/kantik001)).

Include as much detail as possible:

- Description of the vulnerability
- Steps to reproduce
- Affected components (Go server, Python RAG, webapp, Docker config)
- Potential impact
- Suggested fix (if any)

We aim to acknowledge reports within **72 hours** and provide a status update within **7 days**.

## Disclosure policy

- We will confirm receipt and work on a fix.
- We will coordinate disclosure timing with you.
- Credit will be given in the advisory unless you prefer to remain anonymous.

## Security model overview

For deployment and data-flow details, see [docs/en/SECURITY_BRIEF.md](docs/en/SECURITY_BRIEF.md).

### Trust boundaries

| Component | Notes |
|-----------|-------|
| **Go server** | Public API; Telegram HMAC, API keys, admin Basic Auth / OIDC |
| **Python RAG** | Internal service; `/rag/context` has **no auth** — must not be exposed on public networks |
| **PostgreSQL / Chroma / data/** | Client-side storage; stays in your infrastructure |
| **`/metrics`** | Unauthenticated by default — restrict via network policy in production |

### Out of scope for this repository

- Vulnerabilities in third-party LLM providers (OpenRouter, OpenAI, etc.)
- Misconfiguration by deployers (default passwords, exposed admin ports)
- Issues requiring physical access to client infrastructure

We **do** accept reports for:

- Authentication bypass in Go admin or chat APIs
- Cross-tenant data leakage
- Path traversal or unsafe file upload handling
- SSRF via misconfigured URLs
- Secrets logged or exposed in responses

## Secure deployment checklist

Before production:

- [ ] Change default Postgres credentials (`docker-compose.yml` / `DATABASE_URL`)
- [ ] Set strong `ADMIN_PASSWORD` and `ADMIN_SECRET`
- [ ] Do not expose Python RAG port (`5000`) publicly
- [ ] Set `TELEGRAM_AUTH_DISABLED=false` (never in production)
- [ ] Restrict `/metrics` to internal network
- [ ] Configure `CORS_ALLOWED_ORIGINS` explicitly
- [ ] Use OIDC SSO for admin in enterprise deployments ([config/SSO.md](config/SSO.md))

## Dependencies

We track dependency updates via [Dependabot](../.github/dependabot.yml). Report supply-chain concerns through the same private channels above.
