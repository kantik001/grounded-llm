# Local LLM providers (v0.3)

Grounded LLM talks to any **OpenAI-compatible** `/v1/chat/completions` endpoint. Switch providers with env only — no code changes.

| `LLM_PROVIDER` | Default base URL | Typical model env |
|----------------|------------------|-------------------|
| `openai` (default) | `https://openrouter.ai/api` | `LLM_MODEL=openrouter/free` + `LLM_API_KEY` |
| `ollama` | `http://ollama:11434` | `OLLAMA_MODEL=llama3.2` (API key auto `local`) |
| `vllm` | `http://vllm:8000` | `VLLM_MODEL=meta-llama/Meta-Llama-3.1-8B-Instruct` |

Override base URL anytime with `LLM_BASE_URL`.

## Compose profiles

```bash
# Redis + Postgres + stack (always)
docker compose up -d --build

# + Ollama (CPU / Windows-friendly)
docker compose --profile ollama up -d
docker compose exec ollama ollama pull llama3.2
# .env: LLM_PROVIDER=ollama

# + vLLM (NVIDIA GPU)
docker compose --profile vllm up -d
# .env: LLM_PROVIDER=vllm

# Optional: serve through grounded-vllm verify proxy (sibling repo)
# grounded-vllm serve --upstream http://127.0.0.1:8000 --guardrails 127.0.0.1:50052 --port 8001
# .env: LLM_PROVIDER=vllm
#       LLM_BASE_URL=http://127.0.0.1:8001/v1
# See https://github.com/kantik001/grounded-vllm
```

## Caching

| Cache | Key | TTL | Signal |
|-------|-----|-----|--------|
| Embeddings (Python) | `embedding:{md5(text)}:{model}` | 1h | `rag_embedding_cache_hit_total`, Python `/metrics` |
| LLM answers (Go) | `response:{md5(query\|domain\|tenant)}:{model}` | 24h | HTTP `X-Cache: HIT\|MISS` |

## gRPC Retriever

Python exposes `grounded.rag.v1.Retriever` on `:50051` (same container as HTTP RAG).

```bash
grpcurl -plaintext -d '{"query":"How many vacation days?","domain_id":"default","tenant_id":"default","locale":"en","top_k":4}' \
  localhost:50051 grounded.rag.v1.Retriever/Retrieve
```

Auth: metadata `x-rag-service-token` when `RAG_SERVICE_TOKEN` is set.

## Metrics

Go `/metrics` exports `llm_tokens_input_total{tenant,model}`, `llm_tokens_output_total{tenant,model}`, `llm_latency_seconds_*`, `llm_ttft_seconds_*` (stream TTFT), plus response-cache counters. Token counts use provider `usage` when present, otherwise a ~4 chars/token heuristic.
