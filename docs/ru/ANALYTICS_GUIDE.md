**Канон (EN):** [ANALYTICS_GUIDE.md](../en/ANALYTICS_GUIDE.md)

# Analytics Guide — как измерять качество ассистента

Для product owner и platform-инженеров на пилоте или в проде.

Технический API: [config/ANALYTICS.md](../../config/ANALYTICS.md).  
Код: `server/internal/analytics/`, дашборд — `server/internal/store/analytics_*.go`.

---

## Зачем нужна аналитика

Document-grounded ассистенты ломаются предсказуемо:

1. **Retrieval miss** — документ есть, но выбран не тот чанк  
2. **Generation drift** — контекст хороший, формулировка или «изобретённая» деталь плохие  
3. **Verify fail** — числовая галлюцинация при нормальном тексте  
4. **KB gap** — вопрос вне загруженных документов  

Grounded LLM пишет сигналы по каждому пути, чтобы чинить данные / retrieval / промпты точечно.

---

## Ключевые метрики

| Метрика | Где | Если плохо |
|---------|-----|------------|
| **questions_total** | Admin analytics | Низко → adoption; высоко → capacity |
| **rag.verify_pass_rate** | Admin analytics | &lt;75% → промпты, KB, правила verify |
| **rag.soft_fail** | Admin analytics | Высоко → нет доков / плохой chunking; добавить контент или eval |
| **kb_gaps** | Admin analytics | Повторяющиеся темы → статьи KB / FAQ pack |
| **feedback thumbs** | Admin analytics | Негативный кластер → verify fails + citations |
| **Retrieval eval pass** | CI / `run_rag_eval.py` | Регрессия до релиза — блок deploy |

---

## Еженедельный ритуал (15 мин)

1. **Admin → Analytics** (`webapp/admin.html`)  
2. Тренд **verify_pass_rate** (окно 7 дней)  
3. Топ-5 **kb_gaps**  
4. Негативы **feedback**  
5. Подозрение на retrieval → `make eval-retrieval-ci`  

---

## Метрика → продуктовое решение

| Наблюдение | Вероятная причина | Действие |
|------------|-------------------|----------|
| Soft-fail на «VPN» | Нет IT-дока | Pack IT support или загрузить VPN policy |
| Verify fail на датах/суммах | LLM перефразировал числа | Ужесточить промпты; добавить eval case |
| Verify ок, плохой UX-текст | Только generation | Тюнинг промпта (не больше retrieval) |
| Цитаты есть, но не тот файл | Ranking | Эксперимент с reranker; расширить eval |

---

## Шаблон KPI пилота

| KPI | Week 1 | Week 2 | Target |
|-----|--------|--------|--------|
| verify_pass_rate | | | ≥75% |
| soft_fail rate | | | ↓ week over week |
| Avg questions/day | | | ↑ adoption |
| Eval suite pass | | | 100% |

---

## Privacy

Превью вопросов в KB gaps обрезается до **80** символов. Не индексируйте секреты, которые нельзя показать HR-менеджеру в одной строке превью.

---

См. также: [CASE_STUDY_HR_PILOT.md (EN)](../en/CASE_STUDY_HR_PILOT.md) · [SECURITY_BRIEF.md](./SECURITY_BRIEF.md) · [server-oidc-saas-analytics.md](./knowledge-base/server-oidc-saas-analytics.md)
