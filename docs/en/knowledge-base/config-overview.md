# Config directory (`config/`)

**Folder:** `config/` — JSON configs mounted without rebuild (Docker volume `/config`).  
**Readers:** Go (`server/`), Python (`rag/`)  
**Map:** [config/README.md](../../../config/README.md)

---

## Core files

| Path | Reader | Purpose |
|------|--------|---------|
| `domains.json` | Go + Python | Domain catalog, `rag_enabled`, localized `names` |
| `schemas/domains.schema.json` | CI / pytest | Structural validation of `domains.json` |
| `locales/{ru,en}/prompts.json` | Go | Per-domain RAG system prompts + `_platform` rules |
| `locales/{ru,en}/few_shot.json` | Python | Few-shot examples (`locale` in `POST /rag/context`) |
| `locales/{ru,en}/onboarding.json` | Go | Starter question chips in the Web App |
| `locales/{ru,en}/branding.json` | Go | UI titles and disclaimer |
| `examples/*.json.example` | operators | Templates for tenants, quotas, RBAC, OIDC, API keys |

See [config/locales/README.md](../../../config/locales/README.md).

---

## `domains.json`

- **Go:** `GET /domains` returns display names for the request locale
- **Python:** `normalize_domain_id`, Chroma filter by `domain_id` and `tenant_id`

Env: `DOMAINS_CONFIG_PATH`. Schema: `config/schemas/domains.schema.json` (`tests/test_domains_schema.py`).

More: [rag-domains_config.md](./rag-domains_config.md).

---

## Locale bundles

Prompts and UI copy live under `config/locales/{locale}/`.

Go loads all bundles at startup (`initLocaleConfig`). Per request:

- Header `X-Locale` or `Accept-Language`
- Query `?locale=en`
- Telegram `language_code` when using Web App auth

Env: `DEFAULT_LOCALE` (default **`en`**), `LOCALES_ROOT`.

---

## Reload

| Change | Action |
|--------|--------|
| `domains.json` | Go SIGHUP / interval reload; restart Python if needed |
| `locales/*` | Go reload; restart Python for few_shot cache |
| KB files under `data/{tenant}/{domain}/` | Run reindex |

---

## New domain checklist

1. Add entry to `domains.json` (with `names.ru` / `names.en`)
2. Add blocks under both `locales/ru/` and `locales/en/`
3. Place documents in `data/{tenant_id}/{domain_id}/`
4. `python scripts/reindex_rag.py` or `scripts/init_domain.ps1`
