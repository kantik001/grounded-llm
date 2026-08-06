**Канон (EN):** [ECOSYSTEM.md](../en/ECOSYSTEM.md)

# Экосистема — стандарт Grounded и соседние проекты

Grounded LLM — **референсная реализация** открытого стандарта document-grounded ассистентов. Агенты, tool-use и workflow-графы живут в **отдельных репозиториях** и подключаются через публичный API — не внутри ядра стандарта.

См.: [STANDARD_STRATEGY.md](./STANDARD_STRATEGY.md) · [PLATFORM_VISION.md](../../PLATFORM_VISION.md) · [RFC-0001 (EN)](../en/rfcs/RFC-0001-grounded-compatible.md)

---

## Что остаётся в этом репозитории (ядро стандарта)

| Область | Почему здесь |
|---------|----------------|
| **Grounded Spec v1** | Нормативный API и поведение |
| **Conformance CLI** | Проверяемый ярлык «Grounded-compatible» |
| **Качество retrieval** | Eval, benchmark, adversarial gates |
| **Citations + verify** | Ключевые отличия продукта |
| **Vector / hybrid retrieval** | Измеримое качество (dense, BM25+RRF, адаптеры) |
| **Multi-tenant API** | Поверхность для интеграторов |
| **Domain packs** | Единица template marketplace |
| **On-prem deploy** | Docker, Helm, Terraform reference |

Метрика успеха: **новый grounded-ассистент из template &lt;3 дней**, pass rate eval на каждом релизе.

---

## Что уходит в соседние проекты

| Возможность | Отдельный проект | Связь с Grounded |
|-------------|------------------|------------------|
| **ReAct / tool-calling agents** | [grounded-agent](https://github.com/kantik001/grounded-agent) | gRPC Retriever / `POST /rag/context` + MCP Gateway |
| **MCP gateway / registry** | [mcp-gateway](https://github.com/kantik001/mcp-gateway) | Проксирует tools; retrieval остаётся в Grounded |
| **Token-level / remote verify** | [grounded-guardrails](https://github.com/kantik001/grounded-guardrails) | Опц. `GUARDRAILS_MODE=remote` → gRPC `:50052` (`VerifyText`) |
| **Verifiable-generation bench** | [grounded-bench](https://github.com/kantik001/grounded-bench) | Offline NVR / CP / HR / RR; retrieval gates — в `eval/` этого репо |
| **vLLM serving-path verify** | [grounded-vllm](https://github.com/kantik001/grounded-vllm) | OpenAI proxy `:8001` → vLLM `:8000` + guardrails `:50052`; `LLM_BASE_URL` → proxy |
| **Visual workflow builder** | Не планируется | Вне scope ([STANDARD_STRATEGY.md](./STANDARD_STRATEGY.md)) |
| **Обычный chatbot без KB** | Не планируется | Вне scope |

**Правило:** произвольный tool-use / agent graphs **не** входят в Grounded Spec v1 без нового RFC и major bump.

---

## Roadmap (стандарт-first)

### Horizon 1 — референс (сейчас)

| Приоритет | Работа | Pillar |
|-----------|--------|--------|
| ✅ | Hybrid retrieval (BM25 + dense + RRF) | 2, 3 |
| 🔜 | Hybrid modes в Grounded Spec §7 | 1 |
| ✅ | pgvector (`VECTOR_STORE=pgvector`) | 2, 3 |
| 🔜 | Benchmark badge + `bench_report.py` в release | 2 |
| 🔜 | 4-й domain pack с eval | 4 |

### Horizon 2 — platform standard (6–18 мес.)

Conformance режимов retrieval, connector ingest в spec, partner certification, альтернативная реализация проходит conformance.

### Horizon 3 — industry (18+ мес.)

Публичный spec site, «Grounded-compatible» в RFP, [grounded-bench](https://github.com/kantik001/grounded-bench) как цитируемый бенчмарк (**v0 shipped**).

---

## Агент (отдельный репо)

**Repo:** [github.com/kantik001/grounded-agent](https://github.com/kantik001/grounded-agent)

**Scope:** Go ReAct (max 5 steps), tool `retrieve[...]` → Retriever / `/rag/context`, MCP через mcp-gateway, Redis `session:{id}`.

**Non-goals:** замена retrieval/verify Grounded; форк domain packs / eval; ярлык «Grounded-compatible» без conformance против самого Grounded.

---

## См. также

- [GROUNDED_SPEC_v1.md (EN)](../en/spec/GROUNDED_SPEC_v1.md)
- [BENCHMARK.md](./BENCHMARK.md)
- [COMPARISON.md (EN)](../en/COMPARISON.md)
