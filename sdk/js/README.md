# Grounded LLM JavaScript / TypeScript SDK

Minimal client for the public Go REST API (`server/`, typically `:8080`).

```bash
cd sdk/js
npm install
npm run build
```

```ts
import { GroundedClient } from "@grounded-llm/sdk";

const client = new GroundedClient({
  baseUrl: "http://localhost:8080",
  apiKey: process.env.GROUNDED_API_KEY,
  tenantId: "default",
});

const result = await client.chat("How many paid vacation days do employees get?", {
  domainId: "default",
});
console.log(client.lastAssistant(result)?.content);
```

Headers: `X-API-Key`, `X-Tenant-ID`, `X-Locale`. See [QUICKSTART_SDK.md](../../docs/en/QUICKSTART_SDK.md).
