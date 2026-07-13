# Стратегия стандарта — Grounded LLM

Краткая русская версия. **Канон:** [STANDARD_STRATEGY.md (EN)](../en/STANDARD_STRATEGY.md).

---

## Позиционирование

> Открытый стандарт для **ассистентов по внутренним документам** с цитатами, проверкой чисел и измеримым качеством retrieval — на **вашей** инфраструктуре.

**Не конкурируем** с Dify/LangGraph по визуальным workflow. **Конкурируем** по доверию, воспроизводимому качеству и conformance-тестам.

---

## Пять столпов

| # | Столп | Смысл | В репозитории |
|---|-------|-------|---------------|
| 1 | Spec & conformance | Правила + тесты | Spec v1, `python -m conformance` |
| 2 | Quality science | Цифры, не демо | JSONL eval, gate в CI, benchmark |
| 3 | Reference deploy | Повторяемый деплой | Docker, Helm, Terraform |
| 4 | Template marketplace | Рост без форка | HR, IT, Legal FAQ, registry |
| 5 | Governance | Стандарт живёт без автора | RFC, partner cert |

---

## Два пути (совместимы)

| | **A — Open standard** | **B — Product / SaaS** |
|--|-------------------------|-------------------------|
| Фокус | Спека, экосистема | Подписка, удобство |
| Код | MIT core | MIT + optional hosted |
| Сейчас | Launch, conformance | Задел signup/Stripe (фазы 10–11) |

---

## Горизонты

1. **Reference impl** — любой инженер: green conformance за 15 минут  
2. **Platform standard** — интеграторы, коннекторы, packs  
3. **Industry standard** — «Grounded-compatible» в RFP

---

## Связанные документы

- [ROADMAP.md](./ROADMAP.md) · [LAUNCH.md](./LAUNCH.md)
- [GROUNDED_SPEC_v1.md](../en/spec/GROUNDED_SPEC_v1.md) (EN)
