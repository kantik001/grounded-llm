# SDK Quickstart — 5 minutes to a cited answer

Use the Python SDK or CLI against a local Grounded LLM deployment.

## 1. Start the platform

```bash
cp .env.example .env
# LLM_API_KEY=... (OpenAI-compatible). For smoke without LLM bill: LLM_MOCK=true RAG_MOCK=true
# TELEGRAM_AUTH_DISABLED=true

docker compose up -d --build
python scripts/reindex_rag.py
```

## 2. Install the SDK

```bash
pip install -e "sdk/python[dev]"
```

## 3. One-liner CLI

```bash
export GROUNDED_BASE_URL=http://localhost:8080
grounded-llm chat "How many paid vacation days do employees get?" --domain default
```

Expected: answer containing **28** and source filename in stderr when `--show-sources` (default).

## 4. Python integrator snippet

```python
from grounded_llm import GroundedClient

client = GroundedClient("http://localhost:8080", tenant_id="default")
result = client.chat("How many paid vacation days do employees get?", domain_id="default")
print(result.last_assistant_message["content"])
```

Full example: [examples/python/chat_basic.py](../../examples/python/chat_basic.py)

## 5. Production integrators

```python
client = GroundedClient(
    "https://your-host.example.com",
    api_key=os.environ["GROUNDED_API_KEY"],
    tenant_id="acme",
    locale="en",
)
```

Headers map to REST: `X-API-Key`, `X-Tenant-ID`, `X-Locale`. See [API_EXAMPLES.md](./API_EXAMPLES.md).

## 6. Streaming

```bash
grounded-llm chat "Summarize vacation policy" --stream
```

```python
sid = client.create_session("default")
for token in client.stream_message_deltas("How many vacation days?", session_id=sid):
    print(token, end="", flush=True)
```

## See also

- [sdk/python/README.md](../../sdk/python/README.md)
- [OpenAPI](http://localhost:8080/api/v1/openapi.json)
- [PILOT_PLAYBOOK.md](./PILOT_PLAYBOOK.md)
