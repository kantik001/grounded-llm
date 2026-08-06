# Оффер пилота — Grounded LLM (8 недель)

Кратко для заказчика. Полная EN-версия: [PILOT_OFFER.md](../en/PILOT_OFFER.md).  
Чеклист: [PILOT_CHECKLIST.md](../en/PILOT_CHECKLIST.md).

## Суть

Онпрем-ассистент **только по вашим документам**, с **цитатами** и проверкой чисел (Spec verify). Один вертикал (HR или IT), один tenant, **8 недель**, фиксированный scope.

## KPI (пример)

- Recall@k на golden ≥ 90%  
- Цитаты на in-scope ≥ 95%  
- Выдуманные числа — reject  
- Отказ на вопросах вне KB  
- p95 latency — по договорённости  

## День 0

```bash
python scripts/pilot_day0.py --pack hr --tenant pilot-acme
```

Документы → reindex → golden eval → UAT → KPI-отчёт на 8-й неделе.
