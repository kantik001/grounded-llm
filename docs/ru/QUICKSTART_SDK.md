**Канон (EN):** [QUICKSTART_SDK.md](../en/QUICKSTART_SDK.md)

# SDK Quickstart — ответ с цитатами за 5 минут

Python SDK, JS SDK или CLI против локального развёртывания Grounded LLM.

## 1. Поднять платформу

```bash
cp .env.example .env
# LLM_API_KEY=... (OpenAI-compatible) или LLM_PROVIDER=ollama|vllm — см. [LLM_PROVIDERS.md](./LLM_PROVIDERS.md)
# Smoke без счёта LLM: LLM_MOCK=true RAG_MOCK=true
# TELEGRAM_AUTH_DISABLED=true

docker compose up -d --build
python scripts/reindex_rag.py
```

Go API: `http://localhost:8080`.

## 2. Установить SDK

**Python:**

```bash
pip install -e "sdk/python[dev]"
```

**TypeScript / JavaScript:**

```bash
cd sdk/js && npm install && npm run build
```

## 3. One-liner CLI

```bash
export GROUNDED_BASE_URL=http://localhost:8080
grounded-llm chat "How many paid vacation days do employees get?" --domain default
```

Ожидание: в ответе **28**, имя файла источника в stderr при `--show-sources` (по умолчанию).

## 4. Python

```python
from grounded_llm import GroundedClient

client = GroundedClient("http://localhost:8080", tenant_id="default")
result = client.chat("How many paid vacation days do employees get?", domain_id="default")
print(result.last_assistant_message["content"])
```

Полный пример: [sdk/python/examples/chat_basic.py](../../sdk/python/examples/chat_basic.py)

## 5. Production

```python
client = GroundedClient(
    "https://your-host.example.com",
    api_key=os.environ["GROUNDED_API_KEY"],
    tenant_id="acme",
    locale="ru",
)
```

Заголовки REST: `X-API-Key`, `X-Tenant-ID`, `X-Locale`. См. [API_EXAMPLES.md](./API_EXAMPLES.md).

## 6. Streaming

```bash
grounded-llm chat "Summarize vacation policy" --stream
```

```python
sid = client.create_session("default")
for token in client.stream_message_deltas("How many vacation days?", session_id=sid):
    print(token, end="", flush=True)
```

## См. также

- [sdk/python/README.md](../../sdk/python/README.md)
- [sdk/js/README.md](../../sdk/js/README.md)
- [OpenAPI](http://localhost:8080/api/v1/openapi.json)
- [CASE_STUDY_HR_PILOT.md (EN)](../en/CASE_STUDY_HR_PILOT.md)
