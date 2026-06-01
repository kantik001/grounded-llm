# HR Domain Pack (демо / продажи)

**Продукт:** ассистент по HR-политикам и employee handbook  
**Domain ID:** `default` (демо) или `hr` у клиента  
**Locale:** `config/locales/en/` для международных продаж

---

## Что продаём

Внутренний HR-бот с ответами **только из документов компании**:

- Отпуск и планирование  
- Больничные  
- Удалёнка / гибрид  
- Этика и эскалация в HR  

**One-liner для pitch:**  
*Сотрудники получают мгновенные ответы с цитатами из вашего HR handbook — в вашей инфраструктуре.*

---

## Что входит в pack

| Актив | Путь |
|-------|------|
| Demo KB (EN) | `data/default/*_en.txt` |
| Промпты RAG | `config/locales/en/prompts.json` |
| Onboarding | `config/locales/en/onboarding.json` |
| UI branding | `config/locales/en/branding.json` |
| Few-shot | `config/locales/en/few_shot.json` |
| Eval (EN) | `eval/rag_default_en_baseline.jsonl` |
| Сценарий demo | [DEMO_SCRIPT.md](./DEMO_SCRIPT.md) |

Русские демо-документы (`policy_*.txt` без `_en`) — для локали `ru` и русскоязычных пилотов.

---

## Внедрение у клиента (2–5 дней)

1. Запись в `config/domains.json` (domain `hr`).
2. Документы клиента в `data/{tenant}/hr/`.
3. Настройка промптов в `config/locales/en/`.
4. Onboarding-вопросы под темы клиента.
5. Reindex.
6. Eval: `python scripts/run_rag_eval.py --suite default_en`.
7. Пилот по [PILOT_PLAYBOOK.md](../PILOT_PLAYBOOK.md).

---

## Цены (ориентир)

| | USD |
|---|-----|
| Setup pack | $3k–8k или в пилоте |
| Пилот 8 нед | $8k–25k |
| Лицензия / год | $24k–80k |

---

## Не входит (Фаза A)

- Расчёт зарплаты, персональные данные  
- SSO / RBAC — **Фаза B**

---

См. [SECURITY_BRIEF.md](../SECURITY_BRIEF.md), [LOCALE_GUIDE.md](../LOCALE_GUIDE.md).
