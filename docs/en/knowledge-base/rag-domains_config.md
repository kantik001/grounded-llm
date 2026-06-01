# `rag/domains_config.py` — domain catalog

**Source:** `rag/domains_config.py`  
**Config:** `config/domains.json`  
**Used by:** `rag/vector_store.py`, `rag/retrieval.py`, Go (`server/domains.go`), tests

---

## Purpose

Load the **knowledge domains** catalog (workspace / KB identifier).

---

## API

| Function | Description |
|----------|-------------|
| `load_domains_config()` | read JSON, cache by mtime |
| `normalize_domain_id(id)` | trim, lower, existence check |
| `get_domain(id)` | domain metadata |
| `list_domains()` | `{ default_domain, domains }` |
| `default_domain_id()` | usually `"default"` |

---

## `domains.json` format

```json
{
  "default_domain": "default",
  "domains": {
    "default": {
      "name": "Knowledge base",
      "names": { "ru": "<Russian UI label>", "en": "Knowledge base" },
      "emoji": "📚",
      "rag_enabled": true,
      "rag_k": 8
    }
  }
}
```

---

## Domain flags

| Field | Effect |
|-------|--------|
| `rag_enabled: false` | retrieval returns “KB not connected” error |
| `rag_k` | number of fragments (1–20, default 8) |
| `name` / `names` / `name_ru` | UI display (Go uses locale) |
| `ui_hidden` | hide from `GET /domains` |

---

## Env

| Variable | Default path |
|----------|--------------|
| `DOMAINS_CONFIG_PATH` | `config/domains.json` |

In Docker: `/config/domains.json`.

---

## Go mirror

`server/domains.go` — same catalog for `GET /domains` and guards in `domain_guards.go`.

After JSON change: restart Python; Go — SIGHUP or `CONFIG_RELOAD_INTERVAL_SEC`.
