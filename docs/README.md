# Documentation

**Primary language:** English — [`docs/en/`](en/)

| Topic | English |
|-------|---------|
| Platform vision | [PLATFORM_VISION.md](../PLATFORM_VISION.md) (repo root) |
| Architecture | [ARCHITECTURE.md](en/ARCHITECTURE.md) |
| Deploy | [DEPLOY.md](en/DEPLOY.md) |
| LLM providers / Redis / gRPC | [LLM_PROVIDERS.md](en/LLM_PROVIDERS.md) |
| Compatibility | [COMPATIBILITY.md](en/COMPATIBILITY.md) |
| Benchmarks | [BENCHMARK.md](en/BENCHMARK.md) |
| Roadmap (phases 1–11) | [ROADMAP.md](en/ROADMAP.md) |
| Public launch | [LAUNCH.md](en/LAUNCH.md) · [RELEASE.md](en/RELEASE.md) |
| Security | [SECURITY_BRIEF.md](en/SECURITY_BRIEF.md) |
| API examples | [API_EXAMPLES.md](en/API_EXAMPLES.md) |
| Connectors | [CONNECTORS.md](en/CONNECTORS.md) |
| Optional SaaS / billing | [SAAS.md](en/SAAS.md) · [BILLING.md](en/BILLING.md) |
| Ecosystem (standard vs agents) | [ECOSYSTEM.md](en/ECOSYSTEM.md) |
| Guardrails gRPC verify (optional) | [GUARDRAILS.md](en/GUARDRAILS.md) |
| Locale guide | [LOCALE_GUIDE.md](en/LOCALE_GUIDE.md) |
| Knowledge base (deep dives) | [en/knowledge-base/](en/knowledge-base/README.md) |
| Vector stores / hybrid | [VECTOR_STORE.md](en/VECTOR_STORE.md) |
| SDK quickstart | [QUICKSTART_SDK.md](en/QUICKSTART_SDK.md) |
| Analytics | [ANALYTICS_GUIDE.md](en/ANALYTICS_GUIDE.md) |
| Embed widget | [EMBED.md](en/EMBED.md) |
| K8s / Terraform | [K8S_DEPLOY.md](en/K8S_DEPLOY.md) · [TERRAFORM.md](en/TERRAFORM.md) |
| Demo (5 min) | [DEMO.md](en/DEMO.md) |

## Reference templates

| Template | Description |
|----------|-------------|
| [HR domain pack](en/domain-packs/HR.md) | Policy / employee handbook assistant (EN) |
| [IT Support pack](en/domain-packs/IT_SUPPORT.md) | Internal IT helpdesk runbooks (EN) |
| [Legal FAQ pack](en/domain-packs/LEGAL_FAQ.md) | NDA / compliance FAQ (EN) |
| [Demo script](en/domain-packs/DEMO_SCRIPT.md) | 30-minute live demo (HR) |
| [Pack registry](../../packs/registry.yaml) | Official packs (`init_pack.py registry --validate`) |
| [packs/](../packs/) | Official packs + `init_pack.py` scaffold (`pack.yaml` v1) |

Locale bundles and LLM prompts: `config/locales/{en,ru}/`.

## Phase plans (delivery history)

| Phase | Doc |
|-------|-----|
| 4 Spec & trust | [PHASE_4.md](en/PHASE_4.md) |
| 5 Standard publication | [PHASE_5.md](en/PHASE_5.md) |
| 6–8 Ecosystem & connectors | [PHASE_6.md](en/PHASE_6.md) … [PHASE_8.md](en/PHASE_8.md) |
| 9–11 Launch & SaaS | [PHASE_9.md](en/PHASE_9.md) … [PHASE_11.md](en/PHASE_11.md) |

## Russian documentation

[`docs/ru/`](ru/) — **полный зеркальный набор** для изучения продукта (архитектура, деплой, KB, packs, SDK, ops).  
**Канон** для Spec / RFC / PHASE 4–11: [en/](en/).

| RU hub | [ru/README.md](ru/README.md) — путь чтения |
|--------|--------------------------------------------|
| Knowledge base (RU) | [ru/knowledge-base/README.md](ru/knowledge-base/README.md) |
| Standard strategy (RU) | [ru/STANDARD_STRATEGY.md](ru/STANDARD_STRATEGY.md) |

## Phase A deliverables (reference)

| Document | EN | RU |
|----------|----|----|
| Security brief | [SECURITY_BRIEF.md](en/SECURITY_BRIEF.md) | [SECURITY_BRIEF.md](ru/SECURITY_BRIEF.md) |
| HR template | [HR.md](en/domain-packs/HR.md) | [HR.md](ru/domain-packs/HR.md) |
| Case study (pilot template) | [CASE_STUDY_HR_PILOT.md](en/CASE_STUDY_HR_PILOT.md) | — |
| Demo script | [DEMO_SCRIPT.md](en/domain-packs/DEMO_SCRIPT.md) | [DEMO_SCRIPT.md](ru/domain-packs/DEMO_SCRIPT.md) |
