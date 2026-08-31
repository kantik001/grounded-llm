# Документация (русский)

Русская документация покрывает **весь продукт** для изучения и эксплуатации.  
Нормативные тексты стандарта (Spec v1, RFC, фазы 4–11) остаются на английском — ниже есть прямые ссылки.

**EN-канон:** [`docs/en/`](../en/) · **карта репо:** [`PROJECT_STRUCTURE.md`](../../PROJECT_STRUCTURE.md)

---

## С чего начать (путь чтения)

1. [ARCHITECTURE.md](./ARCHITECTURE.md) — слои, порты, поток сообщения  
2. [DEMO.md](./DEMO.md) — 5 минут в чате (HR)  
3. [DEPLOY.md](./DEPLOY.md) — Compose / prod  
4. [knowledge-base/PROJECT_STRUCTURE.md](./knowledge-base/PROJECT_STRUCTURE.md) — папки и пакеты  
5. [knowledge-base/server-overview.md](./knowledge-base/server-overview.md) + [python-api.md](./knowledge-base/python-api.md)  
6. [VECTOR_STORE.md](./VECTOR_STORE.md) · [GUARDRAILS.md](./GUARDRAILS.md) · [ECOSYSTEM.md](./ECOSYSTEM.md)

---

## Обзор платформы

| Тема | Документ |
|------|----------|
| Архитектура | [ARCHITECTURE.md](./ARCHITECTURE.md) |
| Развёртывание | [DEPLOY.md](./DEPLOY.md) · [K8S_DEPLOY.md](./K8S_DEPLOY.md) · [NETWORK_SECURITY.md](./NETWORK_SECURITY.md) |
| LLM / Redis / кэши | [LLM_PROVIDERS.md](./LLM_PROVIDERS.md) |
| Векторные бэкенды | [VECTOR_STORE.md](./VECTOR_STORE.md) |
| Guardrails verify | [GUARDRAILS.md](./GUARDRAILS.md) |
| Экосистема репо | [ECOSYSTEM.md](./ECOSYSTEM.md) |
| Совместимость | [COMPATIBILITY.md](./COMPATIBILITY.md) |
| Бенчмарк (99 кейсов) | [BENCHMARK.md](./BENCHMARK.md) |
| Сравнение с альтернативами | [COMPARISON.md](../en/COMPARISON.md) *(EN)* |

---

## Практика: демо, API, SDK

| Тема | Документ |
|------|----------|
| Демо 5 мин | [DEMO.md](./DEMO.md) |
| Демо-скрипт 30 мин | [domain-packs/DEMO_SCRIPT.md](./domain-packs/DEMO_SCRIPT.md) |
| Примеры API | [API_EXAMPLES.md](./API_EXAMPLES.md) |
| SDK Python / JS | [QUICKSTART_SDK.md](./QUICKSTART_SDK.md) |
| Embed-виджет | [EMBED.md](./EMBED.md) |
| Локали | [LOCALE_GUIDE.md](./LOCALE_GUIDE.md) |

---

## Шаблоны (domain packs)

| Пак | Документ |
|-----|----------|
| HR | [domain-packs/HR.md](./domain-packs/HR.md) |
| IT Support | [domain-packs/IT_SUPPORT.md](./domain-packs/IT_SUPPORT.md) |
| Legal FAQ | [domain-packs/LEGAL_FAQ.md](./domain-packs/LEGAL_FAQ.md) |
| Установка | `python scripts/init_pack.py install hr` · реестр `packs/registry.yaml` |

---

## Эксплуатация и безопасность

| Тема | Документ |
|------|----------|
| Безопасность | [SECURITY_BRIEF.md](./SECURITY_BRIEF.md) |
| Сеть / prod | [NETWORK_SECURITY.md](./NETWORK_SECURITY.md) |
| Бэкап / restore | [BACKUP_RESTORE.md](./BACKUP_RESTORE.md) |
| Purge тенанта | [TENANT_PURGE.md](./TENANT_PURGE.md) |
| Коннекторы | [CONNECTORS.md](./CONNECTORS.md) |
| Ingestion | [INGESTION.md](./INGESTION.md) |
| Аналитика | [ANALYTICS_GUIDE.md](./ANALYTICS_GUIDE.md) |
| SaaS / billing | [SAAS.md](./SAAS.md) · [BILLING.md](./BILLING.md) |
| Launch / roadmap | [LAUNCH.md](./LAUNCH.md) · [ROADMAP.md](./ROADMAP.md) · [STANDARD_STRATEGY.md](./STANDARD_STRATEGY.md) |

---

## База знаний (deep dive по коду)

Полный указатель: **[knowledge-base/README.md](./knowledge-base/README.md)**

| Область | С чего начать |
|---------|----------------|
| Карта + Docker + CI + миграции | [PROJECT_STRUCTURE](./knowledge-base/PROJECT_STRUCTURE.md) · [docker](./knowledge-base/docker-overview.md) · [CI](./knowledge-base/github-ci.yml.md) · [migrations](./knowledge-base/migrations-overview.md) |
| Python RAG | [python-api](./knowledge-base/python-api.md) → [vector](./knowledge-base/rag-vector_store.md) → [retrieval](./knowledge-base/rag-retrieval.md) → [verifier](./knowledge-base/rag-verifier.md) |
| Go server | [server-overview](./knowledge-base/server-overview.md) → chat → RAG → auth → admin |
| OIDC / SaaS / analytics | [server-oidc-saas-analytics.md](./knowledge-base/server-oidc-saas-analytics.md) |
| Качество | [quality-eval-and-rag-logs.md](./knowledge-base/quality-eval-and-rag-logs.md) (**99** кейсов) |

---

## Норматив (только EN)

| Документ | Путь |
|----------|------|
| Grounded Spec v1 | [spec/GROUNDED_SPEC_v1.md](../en/spec/GROUNDED_SPEC_v1.md) |
| RFC-0001 Grounded-compatible | [rfcs/RFC-0001](../en/rfcs/RFC-0001-grounded-compatible.md) |
| Conformance CLI | [conformance/README.md](../../conformance/README.md) |
| Фазы 4–11 (архив delivery) | [PHASE_4.md](../en/PHASE_4.md) … [PHASE_11.md](../en/PHASE_11.md) |
| Trust center / governance | [TRUST_CENTER](../en/TRUST_CENTER.md) · [GOVERNANCE](../en/GOVERNANCE.md) |

Общий указатель репо: [../README.md](../README.md)
