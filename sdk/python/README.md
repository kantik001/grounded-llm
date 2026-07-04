# Grounded LLM Python SDK

Official Python client and CLI for the [Grounded LLM](https://github.com/kantik001/grounded-llm) REST API.

## Install

From the repository (development):

```bash
pip install -e "sdk/python[dev]"
```

From PyPI (when published):

```bash
pip install grounded-llm
```

## Quick usage

```python
from grounded_llm import GroundedClient

client = GroundedClient(
    "http://localhost:8080",
    api_key="your-key",      # optional if TELEGRAM_AUTH_DISABLED=true
    tenant_id="default",
)

session_id = client.create_session(domain_id="default")
result = client.send_message(
    "How many paid vacation days do employees get?",
    session_id=session_id,
    domain_id="default",
)
print(result.last_assistant_message["content"])
print(result.last_assistant_message.get("citations"))
```

## CLI

```bash
export GROUNDED_BASE_URL=http://localhost:8080
grounded-llm health
grounded-llm domains
grounded-llm chat "How many vacation days?" --domain default
grounded-llm chat "Hello" --stream
grounded-llm eval --suite default_en    # requires repo checkout
grounded-llm pack list
grounded-llm pack install hr
```

Environment variables: `GROUNDED_BASE_URL`, `GROUNDED_API_KEY`, `GROUNDED_TENANT_ID`, `GROUNDED_LOCALE`.

## See also

- [API examples](../../docs/en/API_EXAMPLES.md)
- [SDK quickstart](../../docs/en/QUICKSTART_SDK.md)
