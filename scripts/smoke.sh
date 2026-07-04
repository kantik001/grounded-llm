#!/usr/bin/env bash
# API smoke test. Set TELEGRAM_AUTH_DISABLED=true for POST /api/session without initData.
# Optional: LLM_MOCK=true RAG_MOCK=true for full /message path without external services.
set -euo pipefail

BASE_URL="${1:-http://localhost:8080}"
FAILED=0

check() {
  local name="$1" method="$2" path="$3" body="${4:-}"
  local code
  if [[ -n "$body" ]]; then
    code=$(curl -sS -o /tmp/smoke_body.txt -w "%{http_code}" -X "$method" \
      -H "Content-Type: application/json" \
      -d "$body" "${BASE_URL}${path}" || echo "000")
  else
    code=$(curl -sS -o /tmp/smoke_body.txt -w "%{http_code}" -X "$method" \
      "${BASE_URL}${path}" || echo "000")
  fi
  if [[ "$code" =~ ^2 ]]; then
    echo "[OK] $name ($code)"
  else
    echo "[FAIL] $name (HTTP $code)"
    cat /tmp/smoke_body.txt 2>/dev/null || true
    FAILED=$((FAILED + 1))
  fi
}

check_body_contains() {
  local name="$1" needle="$2"
  if grep -qi "$needle" /tmp/smoke_body.txt; then
    echo "[OK] $name (contains '$needle')"
  else
    echo "[FAIL] $name (missing '$needle')"
    cat /tmp/smoke_body.txt 2>/dev/null || true
    FAILED=$((FAILED + 1))
  fi
}

echo "Smoke test: $BASE_URL"

check health GET /health
check ready GET /ready
check metrics GET /api/metrics
check domains GET /api/domains
check branding GET "/api/branding?locale=en"
check session POST /api/session '{"domain_id":"default"}'

SESSION_ID=""
if command -v jq >/dev/null 2>&1; then
  SESSION_ID=$(jq -r '.session_id // empty' /tmp/smoke_body.txt 2>/dev/null || true)
fi
if [[ -z "$SESSION_ID" ]]; then
  SESSION_ID=$(grep -o '"session_id"[[:space:]]*:[[:space:]]*"[^"]*"' /tmp/smoke_body.txt | head -1 | sed 's/.*"\([^"]*\)"$/\1/' || true)
fi

check onboarding GET "/api/onboarding?domain_id=default"

if [[ -n "$SESSION_ID" ]]; then
  echo "[INFO] session_id=$SESSION_ID"
  check message POST /api/message "{\"session_id\":\"${SESSION_ID}\",\"domain_id\":\"default\",\"text\":\"How many paid vacation days do employees get?\"}"
  check_body_contains "message answer" "28"
  check_body_contains "message citations" "filename"
else
  echo "[WARN] Could not parse session_id — skipping /message smoke"
  FAILED=$((FAILED + 1))
fi

if [[ "$FAILED" -gt 0 ]]; then
  echo "Smoke FAILED: $FAILED check(s)"
  exit 1
fi
echo "Smoke PASSED"
