# Locale bundles

Each supported locale has its own folder:

| File | Purpose |
|------|---------|
| `prompts.json` | Per-domain RAG system prompts + `_platform` constraints |
| `branding.json` | Web App UI labels |
| `onboarding.json` | Sample question chips per domain |
| `few_shot.json` | Few-shot examples for Python retrieval |

Supported locales: `ru`, `en`.

Override root: `LOCALES_ROOT` (e.g. `/config/locales` in Docker).

Server: `DEFAULT_LOCALE` env, request headers `X-Locale` / `Accept-Language`, query `?locale=`.

Legacy files in `config/prompts.json` etc. are kept for reference; the Go server loads from `config/locales/` only.
