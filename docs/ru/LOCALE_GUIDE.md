# Как добавить новый язык (locale)

Новые языки подключаются **без правок ядра** Go/Python — только конфиги и одна строка в `supportedLocales`.  
Сейчас: `en`, `ru`.

---

## 1. Папка locale

```text
config/locales/{код}/
  prompts.json
  branding.json
  onboarding.json
  few_shot.json
```

Скопируйте из `config/locales/en/` и переведите строки.

---

## 2. Регистрация в Go

Файл `server/locale.go`:

```go
var supportedLocales = []string{"ru", "en", "de"}
```

При необходимости расширьте `normalizeLocale()` (например `de-DE` → `de`).

---

## 3. Имена доменов

В `config/domains.json` → `names.{код}` для каждого domain.

---

## 4. Документы KB

Файлы в `data/{tenant}/{domain}/`, затем reindex:

```bash
python scripts/reindex_rag.py
```

Модель embeddings мультиязычная, но **лучшее качество** — когда язык вопроса и документов совпадает.

---

## 5. Web App

- Локаль: Telegram, `Accept-Language`, `?locale=`, `X-Locale`  
- UI: `GET /branding?locale=...`  
- Все ключи из `en/branding.json` должны быть в новом locale  

RTL (арабский и т.д.): позже — `dir="rtl"` в `index.html`.

---

## 6. Python RAG

Поле `locale` в `POST /rag/context`. Few-shot из `config/locales/{locale}/few_shot.json`.

Env: `DEFAULT_LOCALE` (по умолчанию `en`).

---

## 7. Чеклист

- [ ] 4 JSON-файла, UTF-8  
- [ ] `go test ./...`  
- [ ] `/branding?locale=...`  
- [ ] Чат с `X-Locale`  
- [ ] Опционально: eval baseline для языка  

---

## 8. Docker

Volume `./config:/config:ro` — после recreate или hot reload конфигов.

---

См. [config/locales/README.md](../../config/locales/README.md), [domain-packs/HR.md](./domain-packs/HR.md).
