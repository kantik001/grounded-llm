# План: Eval RAG (3B) и логи RAG (3C)

**Статус:** частично реализовано — eval-наборы и `scripts/run_rag_eval.py`, логи `[RAG]` в Go. Связка с feedback в отчёте — в плане.  
**Связь:** [../ROADMAP.md](../ROADMAP.md), [server-rag_chat.md](./server-rag_chat.md), [data-pipeline.md](./data-pipeline.md)

---

## Зачем это нужно

Сейчас качество проверяют вручную в чате и unit-тестами верификатора чисел. Для роста до **Middle-уровня продукта** и пилота нужно:

1. **Воспроизводимый eval** — набор вопросов + прогон после каждого reindex/смены модели.
2. **Наблюдаемость** — понять, *почему* ответ плохой (какие чанки, прошёл ли verify, был ли 👎).

---

## Фаза 3B — Eval-набор (30–50 Q&A)

### Цель

Регрессии RAG: не сломать ответы после новых статей, смены `LLM_MODEL`, правки промпта.

### Что подготовить

Файлы (реализовано): `eval/rag_apple_baseline.jsonl` (30 вопросов), `eval/rag_demo_hr_baseline.jsonl` (5). Формат строки JSON:

```json
{
  "crop_id": "apple",
  "question": "Какие признаки парши на листьях?",
  "expect_contains": ["пятн", "парша"],
  "expect_numbers_from_context": false,
  "notes": "эталон вручную после просмотра ответа с текущим индексом"
}
```

Или проще на старте: таблица в CSV / Markdown с колонками: вопрос | допустим ли ответ «нет в материалах» | ключевые слова | эталонный фрагмент.

### Минимум 30–50 вопросов для яблони

| Категория | Доля | Примеры |
|-----------|------|---------|
| disease | ~30% | парша, пятна, лечение |
| fertilizer | ~30% | дозы, подкормка (с цифрами из статей) |
| variety | ~20% | сорта, рентабельность, склон |
| general / edge | ~20% | вопрос вне статей → «нет в материалах» |

Часть вопросов взять из [config/onboarding.json](../../config/onboarding.json), часть — из реальных пользователей пилота.

### Метрики прогона (ручные + полуавто)

| Метрика | Как считать |
|---------|-------------|
| **retrieval hit** | есть ли `success` от `/rag/context` |
| **verify pass rate** | доля ответов без ⚠️ verify |
| **«нет в материалах» accuracy** | на вопросах вне корпуса — не выдумывает |
| **manual score 1–5** | выборочно 10 ответов после прогона |

### Когда гонять

- после **каждого reindex** с новыми статьями;
- после смены **`LLM_MODEL`** или правки `prompts.json` / `rag_chat.go`;
- перед **пилотом** и перед merge крупных PR.

### Реализация

- Скрипт **`scripts/run_rag_eval.py`**: retrieval-режим → `POST /rag/context`, проверка `expect_contains` / `expect_out_of_scope`.
- Отчёты: `eval/results/<timestamp>_<suite>.json`.
- Запуск: `make eval-retrieval` или `python scripts/run_rag_eval.py --suite all` (нужен classifier на `:5000`).
- См. [eval/README.md](../../eval/README.md), [DEPLOY.md](../DEPLOY.md).

**В плане:** full-режим с LLM через Go; порог pass rate в CI.

---

## Фаза 3C — Логи RAG (наблюдаемость)

### Цель

Разбор плохих ответов и связка с feedback 👍/👎 без логирования тела LLM (политика 1C в ROADMAP).

### Что логируется (Go, `rag_log.go`)

На каждый текстовый ответ в `answerWithRAG` (из `handleTextMessage` / `handleChat`):

| Поле | Пример |
|------|--------|
| `session_id` | hex |
| `message_id` | после INSERT assistant |
| `crop_id` | apple |
| `question` | до 120 символов (обрезка) |
| `crop_id` | apple / demo_hr / … |
| `session_id` | из чата (пусто для `/chat`) |
| `fragments` | число фрагментов из RAG |
| `verify_pass` | true/false |
| `verify_reason` | текст при fail |
| `soft_fail` | RAG не нашёл материалы |

Пример в логах сервера: `[RAG] crop_id=apple session_id=… fragments=4 verify_pass=true …`

**Не логируется:** полный промпт и тело LLM.

### Куда писать дальше (план)

| Вариант | Плюсы |
|---------|--------|
| **Postgres** `rag_query_log` | SQL + связь с `message_id` и feedback |
| **analytics_events** | расширить `event_type: rag_trace` |

Сейчас: **stdout** Go (`docker compose logs server`).

Связка с feedback:

```sql
-- идея: message_id уже в message_feedback
SELECT m.content, mf.rating, /* будущие поля trace */
FROM messages m
JOIN message_feedback mf ON mf.message_id = m.id
WHERE mf.rating = -1;
```

### Использование

1. Пользователь нажал 👎 → найти `message_id` → посмотреть фрагменты и verify.
2. Массово: топ-10 вопросов с `verify_pass = false` за неделю.
3. Дополнение eval-набора реальными провальными вопросами.

---

## Порядок внедрения (рекомендация)

```mermaid
flowchart LR
    A[16 статей в data/apple] --> B[eval 30 вопросов + run_rag_eval]
    B --> C[Ручной прогон в чате + feedback]
    C --> D[Логи 3C в БД опционально]
    F --> G[Пилот пользователей]
```

1. Сначала **контент** ([data-pipeline.md](./data-pipeline.md)).  
2. Параллельно набросать **10–15 eval-вопросов**.  
3. После стабильного RAG — **логи 3C** (проще отладка).  
4. Довести eval до **30–50** перед пилотом.

---

## Чеклист «готово к пилоту» (качество)

- [x] ≥16 статей в `data/apple/`, reindex после добавления  
- [x] Eval-набор 30 вопросов (`rag_apple_baseline.jsonl`), скрипт прогона  
- [x] Retrieval baseline apple 30/30 (после правок KB; отчёты локально в `eval/results/`)  
- [ ] Verify pass rate известен (retrieval + выборочно full LLM)  
- [x] Логи `[RAG]` в Go (fragments + verify + session_id)  
- [ ] 5+ реальных пользователей, feedback разбирается раз в неделю  

---

## Обновление ROADMAP

См. секции **3B** и **3C** в [../ROADMAP.md](../ROADMAP.md).

---

## Краткий итог

**3B Eval** — эталонные вопросы и регрессии качества RAG. **3C Логи** — трассировка от вопроса до verify и связь с 👍/👎. Оба пункта — следующий шаг после наполнения `data/apple/` и стабильного compose, без блокера для чтения текущего кода.
