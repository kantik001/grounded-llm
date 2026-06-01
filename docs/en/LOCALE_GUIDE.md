# Adding a New Locale

Grounded LLM supports **pluggable locales** without changing Go or Python core code.  
Supported today: `en`, `ru`. Adding `de`, `ar`, etc. follows the same pattern.

---

## 1. Create locale folder

```text
config/locales/{code}/
  prompts.json
  branding.json
  onboarding.json
  few_shot.json
```

Copy from `config/locales/en/` and translate all string values.

---

## 2. Register locale in Go

Edit `server/locale.go`:

```go
var supportedLocales = []string{"ru", "en", "de"}  // add code
```

Extend `normalizeLocale()` if you need BCP-47 variants (e.g. `de-DE` → `de`).

---

## 3. Domain display names

In `config/domains.json`, add names per locale:

```json
"names": {
  "en": "Knowledge base",
  "ru": "База знаний",
  "de": "Wissensdatenbank"
}
```

---

## 4. Knowledge base documents

Place translated (or bilingual) files under:

```text
data/{tenant_id}/{domain_id}/
```

Reindex after adding files: `python scripts/reindex_rag.py`.

Embeddings use `intfloat/multilingual-e5-small` — cross-lingual retrieval works, but **best quality** when questions and documents share the same language.

---

## 5. Web App

- User locale: Telegram `language_code`, `Accept-Language`, `?locale=`, or `X-Locale` header  
- UI strings: `GET /branding?locale={code}`  
- Extend `config/locales/{code}/branding.json` with all keys from the English file  

For RTL languages (e.g. Arabic): set `dir="rtl"` on `<html>` in `webapp/index.html` when locale is RTL (future webapp enhancement).

---

## 6. Python RAG

Pass `locale` in `POST /rag/context` JSON. Few-shot examples load from `config/locales/{locale}/few_shot.json`.

Default: `DEFAULT_LOCALE` env (default `en`).

---

## 7. Verify checklist

- [ ] All four JSON files valid UTF-8  
- [ ] `go test ./...` in `server/`  
- [ ] `pytest tests/test_eval_baseline.py`  
- [ ] Manual: `GET /branding?locale={code}`  
- [ ] Manual: chat with `X-Locale: {code}`  
- [ ] Optional: `eval/rag_{domain}_{code}_baseline.jsonl`

---

## 8. Docker

Mount `./config:/config:ro` — new locale folders appear after container recreate or config reload (`CONFIG_RELOAD_INTERVAL_SEC` on Go server).

---

See also: [config/locales/README.md](../../config/locales/README.md), [domain-packs/HR.md](./domain-packs/HR.md).
