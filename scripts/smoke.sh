#!/usr/bin/env bash
# Smoke-тест API. TELEGRAM_AUTH_DISABLED=true для POST /api/session без initData.
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
    FAILED=$((FAILED + 1))
  fi
}

echo "Smoke test: $BASE_URL"

check health GET /health
check crops GET /api/crops
check session POST /api/session '{"crop_id":"apple"}'
check onboarding GET "/api/onboarding?crop_id=apple"

if [[ "$FAILED" -gt 0 ]]; then
  echo "Smoke FAILED: $FAILED check(s)"
  exit 1
fi
echo "Smoke PASSED"
