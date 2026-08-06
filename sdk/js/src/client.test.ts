import assert from "node:assert/strict";
import { describe, it } from "node:test";
import { GroundedClient, GroundedAPIError } from "./client.ts";

describe("GroundedClient", () => {
  it("builds versioned URLs and sends tenant headers", async () => {
    const calls: Array<{ url: string; headers: HeadersInit | undefined }> = [];
    const client = new GroundedClient({
      baseUrl: "http://example.test",
      apiKey: "gk_test",
      tenantId: "acme",
      locale: "en",
      fetch: async (input, init) => {
        calls.push({ url: String(input), headers: init?.headers });
        return new Response(JSON.stringify({ success: true, session_id: "s1" }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      },
    });
    const sid = await client.createSession("default");
    assert.equal(sid, "s1");
    assert.equal(calls[0]?.url, "http://example.test/api/v1/session");
    const h = calls[0]?.headers as Record<string, string>;
    assert.equal(h["X-Tenant-ID"], "acme");
    assert.equal(h["X-API-Key"], "gk_test");
  });

  it("raises GroundedAPIError on success:false", async () => {
    const client = new GroundedClient({
      fetch: async () =>
        new Response(JSON.stringify({ success: false, error: "boom" }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        }),
    });
    await assert.rejects(() => client.health(), (err: unknown) => {
      assert.ok(err instanceof GroundedAPIError);
      assert.equal((err as GroundedAPIError).message, "boom");
      return true;
    });
  });
});
