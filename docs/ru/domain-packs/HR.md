# HR Domain Pack (шаблон)

**Сценарий:** внутренний HR-ассистент и employee handbook  
**Domain ID:** `default` (демо) или свой slug, например `hr`  
**Locale:** `config/locales/en/` (или `ru` для русской локали)

Это **референсный шаблон** Grounded LLM. Скопируйте и адаптируйте, чтобы развернуть document-grounded ассистента за несколько дней.

---

## Что даёт шаблон

Готовый **policy Q&A** с ответами только из документов компании:

- Отпуск и планирование  
- Больничные  
- Удалёнка / гибрид  
- Этика и эскалация в HR  

**One-liner:**

> Сотрудники получают мгновенные ответы с цитатами из handbook — в вашей инфраструктуре.

---

## Состав pack

| Актив | Путь |
|-------|------|
| Demo KB (EN) | `data/default/*_en.txt` |
| Промпты RAG | `config/locales/en/prompts.json` |
| Onboarding | `config/locales/en/onboarding.json` |
| UI branding | `config/locales/en/branding.json` |
| Few-shot | `config/locales/en/few_shot.json` |
| Eval (EN) | `eval/rag_default_en_baseline.jsonl` |
| Манифест pack | `packs/hr/pack.yaml` |
| Сценарий demo | [DEMO_SCRIPT.md](./DEMO_SCRIPT.md) |

Русские демо-документы (`policy_*.txt` без `_en`) — для локали `ru`.

---

## Развёртывание (2–5 дней)

**Рекомендуется — установка из pack:**

```bash
python scripts/init_pack.py install hr
python scripts/reindex_rag.py
python scripts/run_rag_eval.py --suite default_en
```

**Вручную:**

1. Запись в `config/domains.json` (domain `hr`).
2. Документы клиента в `data/{tenant}/hr/`.
3. Настройка `config/locales/en/prompts.json` и `branding.json`.
4. Onboarding-вопросы под темы клиента.
5. `python scripts/reindex_rag.py` или `POST /admin/reindex`.
6. Eval: `python scripts/run_rag_eval.py --suite default_en`.
7. Согласование с IT: [SECURITY_BRIEF.md](../SECURITY_BRIEF.md).

---

## Вне scope (сегодня)

- Расчёт зарплаты, персональные данные  
- SSO / RBAC — см. [roadmap Phase B](../ROADMAP.md)  
- Юридические консультации — disclaimer в branding bundles  

---

## Связанные документы

- [PLATFORM_VISION.md](../../../PLATFORM_VISION.md)  
- [LOCALE_GUIDE.md](../LOCALE_GUIDE.md)  
- [domain-pack-template/](../../../domain-pack-template/)  
- [CASE_STUDY_HR_PILOT.md](../../en/CASE_STUDY_HR_PILOT.md) — шаблон KPI для пилота  
