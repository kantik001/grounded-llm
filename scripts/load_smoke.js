// k6 load smoke — chat session + message against a running server (mock LLM/RAG OK).
// Usage:
//   k6 run -e BASE_URL=http://127.0.0.1:8080 scripts/load_smoke.js
// Requires TELEGRAM_AUTH_DISABLED=true (or API key) on the target.

import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  vus: Number(__ENV.VUS || 20),
  duration: __ENV.DURATION || "30s",
  thresholds: {
    http_req_failed: ["rate<0.05"],
    http_req_duration: ["p(95)<3000"],
  },
};

const BASE = (__ENV.BASE_URL || "http://127.0.0.1:8080").replace(/\/$/, "");

export default function () {
  const health = http.get(`${BASE}/health`);
  check(health, { "health 200": (r) => r.status === 200 });

  const session = http.post(
    `${BASE}/api/session`,
    JSON.stringify({ domain_id: "default" }),
    { headers: { "Content-Type": "application/json" } }
  );
  check(session, { "session 200": (r) => r.status === 200 });
  let sid = "";
  try {
    sid = session.json("session_id") || "";
  } catch (_) {}
  if (!sid) {
    sleep(0.5);
    return;
  }

  const msg = http.post(
    `${BASE}/api/message`,
    JSON.stringify({
      session_id: sid,
      domain_id: "default",
      text: "How many paid vacation days do employees get?",
    }),
    { headers: { "Content-Type": "application/json" } }
  );
  check(msg, { "message 200": (r) => r.status === 200 });
  sleep(0.3);
}
