**Канон (EN):** [BENCHMARK.md](../en/BENCHMARK.md)

# Grounded Benchmark

Публичные метрики качества референса и заявлений **Grounded-compatible** ([RFC-0001 EN](../en/rfcs/RFC-0001-grounded-compatible.md)).

Зачем: каждый релиз — измеримый retrieval pass rate, а не «кажется, что отвечает лучше».

---

## Suites

| Suite | Файл | Кейсы | Что меряет |
|-------|------|-------|------------|
| EN HR baseline | `eval/rag_default_en_baseline.jsonl` | 21 | Retrieval (+ paraphrase) |
| RU legacy | `eval/rag_default_baseline.jsonl` | 12 | Retrieval accuracy |
| IT support | `eval/rag_it_support_baseline.jsonl` | 16 | Cross-template retrieval |
| Adversarial | `eval/rag_adversarial_baseline.jsonl` | 30 | Неверные числа, cross-domain, injection |
| Hybrid | `eval/rag_hybrid_baseline.jsonl` | 7 | BM25+RRF (`RAG_RETRIEVAL_MODE=hybrid`) |
| Legal FAQ | `eval/rag_legal_faq_baseline.jsonl` | 13 | Cross-template retrieval |
| **Retrieval total** | — | **99** | Все JSONL выше (без E2E) |
| Adversarial E2E | `eval/rag_adversarial_e2e.jsonl` | 5 | Полный путь `/message` (mock / staging) |

---

## Локальный прогон

```bash
# Только retrieval (Python RAG :5000)
export PYTHON_RAG_URL=http://localhost:5000/rag/context
python scripts/run_rag_eval.py --suite all

python scripts/bench_report.py

# E2E adversarial (Go с LLM_MOCK + RAG_MOCK, :8080)
python scripts/run_adversarial_e2e.py --base-url http://127.0.0.1:8080
```

---

## CI gates

| Job | Suites |
|-----|--------|
| `eval-retrieval-gate` | Все retrieval JSONL (включая adversarial) — **99** кейсов |
| `smoke-api` | Adversarial E2E (mock) |

---

## Latency провайдеров (иллюстративно)

| Provider | Команда | Mean (3 asks) | p95 | Hardware |
|----------|---------|---------------|-----|----------|
| OpenRouter / OpenAI-compatible | `LLM_PROVIDER=openai` | ~1.8s | ~2.4s | Cloud RTT |
| Ollama `llama3.2` | `--profile ollama` | ~4.5s | ~6.0s | CPU |
| vLLM Llama-3.1-8B | `--profile vllm` | ~0.7s | ~0.9s | NVIDIA |

Процедура: reindex → 3 DEMO-вопроса → wall-clock или `llm_latency_seconds_*` / `llm_ttft_seconds_*` с `/metrics`. Повтор с Redis: второй идентичный ask → `X-Cache: HIT`.

---

## Отчёт в релизе

```bash
python scripts/bench_report.py --write eval/results/latest_bench.json
```

Пример: `"retrieval_total": {"passed": 99, "total": 99, "pass_rate": 1.0}`.

Публичные метрики **verifiable-generation** (NVR / CP / HR / RR) — sibling [grounded-bench](https://github.com/kantik001/grounded-bench) (offline, 280 cases). Этот документ — retrieval / release-bench для grounded-llm.

---

## См. также

- [eval/README.md](../../eval/README.md)
- [STANDARD_STRATEGY.md](./STANDARD_STRATEGY.md)
- [RFC-0001 (EN)](../en/rfcs/RFC-0001-grounded-compatible.md)
