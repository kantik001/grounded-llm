# Публичный launch — чеклист

Краткая русская версия. **Канон:** [LAUNCH.md (EN)](../en/LAUNCH.md).

---

## Перед launch (инженерия)

- [ ] CI зелёный на `main` (eval-retrieval-gate, smoke-api, conformance)
- [ ] Тег релиза `v0.3.0` — [RELEASE.md (EN)](../en/RELEASE.md)
- [ ] `python scripts/init_pack.py registry --validate`
- [ ] `python scripts/build_site_data.py` перед Pages
- [ ] Secret scan, README + demo GIF
- [ ] В публичном дереве нет sales/playbook и неопубликованных черновиков dev.to (см. `.gitignore`)

---

## Сделать репозиторий публичным

1. GitHub → Settings → Visibility → **Public**
2. Pages → Source: **GitHub Actions**
3. Workflow **Deploy site** (вручную; на Free нужен public repo)
4. Проверить: `https://<user>.github.io/grounded-llm/`

---

## Каналы первой волны

| Канал | Угол |
|-------|------|
| dev.to | Платформенная статья — **публикуется на dev.to**, не в репозитории до выхода |
| Hacker News | Show HN — spec + conformance + on-prem |
| Reddit | r/selfhosted, r/MachineLearning |
| LinkedIn / X | Короткое demo + ссылка на spec |

**Одна фраза:** open standard for cited, verified document assistants — Spec v1, conformance CLI, on-prem templates.

---

## После launch

- Issues / Discussions
- [PARTNER_CERTIFICATION.md (EN)](../en/PARTNER_CERTIFICATION.md)
- [GOVERNANCE.md (EN)](../en/GOVERNANCE.md)
