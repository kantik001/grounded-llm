# Legacy flat locale files

These files lived at `config/*.json` before per-locale bundles under `config/locales/{ru,en}/`.

**Runtime does not load this directory.** Go and Python read only `LOCALES_ROOT` (default `config/locales`).

Kept for historical diff / migration reference. Prefer editing `config/locales/` only. Safe to delete in a later cleanup once no forks depend on the old paths.
