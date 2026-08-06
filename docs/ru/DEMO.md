**Канон (EN):** [DEMO.md](../en/DEMO.md)

# Демо за 5 минут (HR pack)

Clone → up → ask. Показывает **ответы с цитатами** из default HR KB и честный отказ, когда ответа нет в документах.

Длинный sales-сценарий: [domain-packs/DEMO_SCRIPT.md](./domain-packs/DEMO_SCRIPT.md).

---

## 1. Старт

```bash
git clone https://github.com/kantik001/grounded-llm.git
cd grounded-llm
cp .env.example .env
```

В `.env` минимум:

```env
TELEGRAM_AUTH_DISABLED=true
LLM_API_KEY=sk-...          # OpenAI-compatible
# Опциональный smoke без реального LLM:
# LLM_MOCK=true
# RAG_MOCK=true
```

```bash
docker compose up -d --build
python scripts/reindex_rag.py
```

UI: **http://localhost/** · API: `http://localhost:8080`.

| Проверка | Ожидание |
|----------|----------|
| `curl -sf http://localhost:8080/health` | `200` |
| Web UI | Чат для domain `default` |

---

## 2. Вопросы in-scope

В web chat. В ответе — блок **Sources** с именем policy-файла.

| # | Вопрос | Ожидание в ответе |
|---|--------|-------------------|
| 1 | How many paid vacation days do employees get? | **28** calendar days |
| 2 | How far in advance must vacation be planned in HR Portal? | **14** days |
| 3 | By what time must I notify my manager on the first sick day? | **10:00** |
| 4 | How many remote work days per week are allowed? | **2** days |
| 5 | Which days are recommended in office for team sync? | **Tuesday** и **Thursday** |
| 6 | What are core messenger availability hours? | **09:00–18:00** |

Те же кейсы в CI: suite `default_en` (`packs/hr/eval.jsonl`).

---

## 3. Out-of-scope (trust)

| Вопрос | Ожидание |
|--------|----------|
| What is the CEO salary on the Moon in 2099? | Без выдуманной зарплаты; отказ «нет в KB» |
| How many vacation days for employees on Mars? | То же — без галлюцинации |

---

## 4. API one-liner (опционально)

При `TELEGRAM_AUTH_DISABLED=true`:

```bash
curl -sS -X POST http://127.0.0.1:8080/api/session \
  -H "Content-Type: application/json" \
  -d '{"domain_id":"default"}'

curl -sS -X POST http://127.0.0.1:8080/api/message \
  -H "Content-Type: application/json" \
  -d '{"session_id":"SESSION_ID","domain_id":"default","text":"How many paid vacation days do employees get?"}'
```

Или: `grounded-llm chat "How many vacation days?" --domain default` после `pip install -e "sdk/python[dev]"` — [QUICKSTART_SDK.md](./QUICKSTART_SDK.md).

---

## 5. Если что-то не так

| Симптом | Фикс |
|---------|------|
| Пустые / размытые ответы | `python scripts/reindex_rag.py`; `docker compose logs python` |
| 401 в браузере | `TELEGRAM_AUTH_DISABLED=true` и рестарт server |
| Ошибки LLM | Валидный `LLM_API_KEY` или `LLM_MOCK=true` для структурного smoke |
| Не тот язык UI | `DEFAULT_LOCALE=ru` или `?locale=ru` |

---

## См. также

- [README Quick start](../../README.md#quick-start)
- [HR pack](./domain-packs/HR.md) · [DEMO_SCRIPT.md](./domain-packs/DEMO_SCRIPT.md)
- [BENCHMARK.md](./BENCHMARK.md)
