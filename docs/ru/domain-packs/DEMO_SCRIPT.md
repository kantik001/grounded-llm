# Сценарий HR-demo (30 минут)

> **Рынок РФ:** используйте [DEMO_SCRIPT_RU.md](./DEMO_SCRIPT_RU.md) и [SALES_ONE_PAGER_RF.md](../SALES_ONE_PAGER_RF.md).

Домен **default**, локаль **en** (`?locale=en` или `DEFAULT_LOCALE=en`).

---

## Подготовка

```bash
docker compose up -d
python scripts/reindex_rag.py
# Для браузера без Telegram: TELEGRAM_AUTH_DISABLED=true
```

Проверка eval (опционально):

```bash
pip install requests
python scripts/run_rag_eval.py --suite default_en
```

---

## Вступление (3 мин)

> «Сотрудники каждую неделю задают HR одни и те же вопросы. ChatGPT для внутренних политик нельзя. Grounded LLM отвечает **только из ваших документов**, с источниками, **в вашей** инфраструктуре.»

---

## Вопросы в чате (10 мин)

Показывайте **ответ + блок Sources**.

| # | Вопрос (EN) | Ожидаемый факт |
|---|-------------|----------------|
| 1 | How many paid vacation days do employees get? | 28 days |
| 2 | How far in advance must vacation be planned in HR Portal? | 14 days |
| 3 | How many vacation days can be carried over? | 14 days max |
| 4 | By what time notify manager on first sick day? | 10:00 |
| 5 | Within how many days submit sick note to HR? | 3 working days |
| 6 | How many remote days per week? | 2 |
| 7 | Recommended in-office days? | Tuesday, Thursday |
| 8 | Messenger availability hours? | 09:00–18:00 |

**Вне базы (доверие):**

- CEO salary on the Moon in 2099 → честное «not in knowledge base»  
- Vacation days on Mars → без галлюцинаций  

---

## Доверие (5 мин)

- Цитата с именем файла  
- Verify для цифр  
- Данные не идут на обучение публичных моделей  

---

## Админка (5 мин)

Upload `.txt` → Reindex → вопрос по новому файлу.

---

## Закрытие (2 min)

Пилот 8 недель, KPI, лицензия. IT — [SECURITY_BRIEF.md](../SECURITY_BRIEF.md).

---

Полный EN-скрипт: [DEMO_SCRIPT.md](../en/domain-packs/DEMO_SCRIPT.md).
