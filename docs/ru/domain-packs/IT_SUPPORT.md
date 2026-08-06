**Канон (EN):** [IT_SUPPORT.md](../en/domain-packs/IT_SUPPORT.md)

# IT Support Template Pack (референс)

**Сценарий:** внутренний IT helpdesk Q&A (пароль, VPN, железо, почта, SLA)  
**Domain ID:** `it_support`  
**Locale:** `config/locales/en/` (и/или `ru`)  
**KB:** `data/default/it_support/`

Второй **официальный шаблон** Grounded LLM — рядом или вместо [HR](./HR.md).

---

## Что даёт шаблон

Ответы только из IT-runbooks:

- Сброс пароля и lockout  
- VPN  
- Laptop / hardware  
- Phishing и лимиты почты  
- Часы IT Portal и ticket SLA (P1–P3)

**One-liner:**

> Сотрудники получают цитируемые ответы из ваших IT-runbooks — в вашей инфраструктуре, с измеримым retrieval.

---

## Состав

| Актив | Путь |
|-------|------|
| Demo KB | `data/default/it_support/*.txt` |
| Domain | `config/domains.json` → `it_support` |
| RAG prompts | `config/locales/en/prompts.json` → `it_support` |
| Onboarding | `config/locales/en/onboarding.json` → `it_support` |
| Few-shot | `config/locales/en/few_shot.json` → `it_support` |
| Eval | `eval/rag_it_support_baseline.jsonl` (**16** cases) |
| Manifest | `packs/it_support/pack.yaml` |

---

## Деплой (2–5 дней)

```bash
python scripts/init_pack.py install it_support
python scripts/reindex_rag.py
python scripts/run_rag_eval.py --suite it_support
```

**Вручную:** domain в `domains.json` → свои runbooks в `data/{tenant}/it_support/` → тюнинг prompts/onboarding → reindex → eval → [SECURITY_BRIEF.md](../SECURITY_BRIEF.md) с IT security.

Legacy scaffold: `./scripts/init_domain.sh it_support default`

---

## Примеры eval

| Вопрос | Факт |
|--------|------|
| How long is a password reset link valid? | 24 hours |
| Which VPN client is approved? | GlobalProtect |
| P1 initial response time? | 1 hour |
| Where to report phishing? | security@company.com |

Полный suite: `eval/rag_it_support_baseline.jsonl`.

---

## Вне scope

- Создание тикетов в ServiceNow/Jira (кастомный SOW)  
- Автоматический сброс пароля  
- Настройка SSO — roadmap платформы  

---

## См. также

- [HR.md](./HR.md) · [LEGAL_FAQ.md](./LEGAL_FAQ.md)  
- [packs/README.md](../../../packs/README.md)  
- [PLATFORM_VISION.md](../../../PLATFORM_VISION.md)  
