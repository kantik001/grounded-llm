**Канон (EN):** [LLM_PROVIDERS.md](../en/LLM_PROVIDERS.md)

# Локальные LLM-провайдеры (v0.3)

Grounded LLM ходит в любой **OpenAI-compatible** endpoint `/v1/chat/completions`. Провайдер переключается только через env — без правок кода.

Зачем: один пайплайн RAG + verify; модель — операторский выбор (облако, Ollama на CPU, vLLM на GPU).

| `LLM_PROVIDER` | Base URL по умолчанию | Типичная модель |
|----------------|----------------------|-----------------|
| `openai` (default) | `https://openrouter.ai/api` | `LLM_MODEL=openrouter/free` + `LLM_API_KEY` |
| `ollama` | `http://ollama:11434` | `OLLAMA_MODEL=llama3.2` (ключ авто `local`) |
| `vllm` | `http://vllm:8000` | `VLLM_MODEL=meta-llama/Meta-Llama-3.1-8B-Instruct` |

Переопределение: `LLM_BASE_URL` в любой момент.

## Compose profiles

```bash
# Redis + Postgres + стек
docker compose up -d --build

# + Ollama (CPU / удобно на Windows)
docker compose --profile ollama up -d
docker compose exec ollama ollama pull llama3.2
# .env: LLM_PROVIDER=ollama

# + vLLM (NVIDIA GPU)
docker compose --profile vllm up -d
# .env: LLM_PROVIDER=vllm

# Опционально: verify-proxy из sibling-репо grounded-vllm
# grounded-vllm serve --upstream http://127.0.0.1:8000 --guardrails 127.0.0.1:50052 --port 8001
# .env: LLM_PROVIDER=vllm
#       LLM_BASE_URL=http://127.0.0.1:8001/v1
# https://github.com/kantik001/grounded-vllm
```

## Кэши (Redis)

| Кэш | Ключ | TTL | Сигнал |
|-----|------|-----|--------|
| Embeddings (Python) | `embedding:{md5(text)}:{model}` | 1ч | `rag_embedding_cache_hit_total`, Python `/metrics` |
| LLM answers (Go) | `response:{md5(query\|domain\|tenant)}:{model}` | 24ч | HTTP `X-Cache: HIT\|MISS` |

Нужен `REDIS_URL` (в Compose Redis на `:6379`).

## gRPC Retriever

Python отдаёт `grounded.rag.v1.Retriever` на **`:50051`** (тот же контейнер, что HTTP RAG `:5000`).

```bash
grpcurl -plaintext -d '{"query":"How many vacation days?","domain_id":"default","tenant_id":"default","locale":"en","top_k":4}' \
  localhost:50051 grounded.rag.v1.Retriever/Retrieve
```

Auth: metadata `x-rag-service-token`, если задан `RAG_SERVICE_TOKEN`.

## Метрики

Go `/metrics` (`:8080`): `llm_tokens_input_total{tenant,model}`, `llm_tokens_output_total{tenant,model}`, `llm_latency_seconds_*`, `llm_ttft_seconds_*` (stream TTFT), плюс счётчики response-cache.

Токены — из `usage` провайдера; иначе эвристика ~4 символа / token (`internal/metrics`).

## См. также

- [GUARDRAILS.md](./GUARDRAILS.md) — remote verify после LLM  
- [VECTOR_STORE.md](./VECTOR_STORE.md) — кэш эмбеддингов и hybrid  
