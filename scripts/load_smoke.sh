#!/usr/bin/env bash
# Concurrent load smoke without k6 (CI-friendly). Uses curl + background jobs.
# Usage: bash scripts/load_smoke.sh [BASE_URL] [CONCURRENCY] [ROUNDS]
set -euo pipefail

BASE_URL="${1:-http://127.0.0.1:8080}"
CONCURRENCY="${2:-20}"
ROUNDS="${3:-3}"
FAILED=0

one_flow() {
  local id="$1"
  local code
  code=$(curl -sS -o /tmp/load_health_"$id".txt -w "%{http_code}" "${BASE_URL}/health" || echo 000)
  if [[ ! "$code" =~ ^2 ]]; then
    echo "[FAIL] health #$id HTTP $code"
    return 1
  fi
  code=$(curl -sS -o /tmp/load_session_"$id".txt -w "%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d '{"domain_id":"default"}' "${BASE_URL}/api/session" || echo 000)
  if [[ ! "$code" =~ ^2 ]]; then
    echo "[FAIL] session #$id HTTP $code"
    return 1
  fi
  local sid
  sid=$(grep -o '"session_id"[[:space:]]*:[[:space:]]*"[^"]*"' /tmp/load_session_"$id".txt | head -1 | sed 's/.*"\([^"]*\)"$/\1/' || true)
  if [[ -z "$sid" ]]; then
    echo "[FAIL] session #$id: no session_id"
    return 1
  fi
  code=$(curl -sS -o /tmp/load_msg_"$id".txt -w "%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d "{\"session_id\":\"${sid}\",\"domain_id\":\"default\",\"text\":\"How many paid vacation days?\"}" \
    "${BASE_URL}/api/message" || echo 000)
  if [[ ! "$code" =~ ^2 ]]; then
    echo "[FAIL] message #$id HTTP $code"
    return 1
  fi
  return 0
}

echo "Load smoke: ${BASE_URL} concurrency=${CONCURRENCY} rounds=${ROUNDS}"
for round in $(seq 1 "$ROUNDS"); do
  echo "--- round $round ---"
  pids=()
  for i in $(seq 1 "$CONCURRENCY"); do
    one_flow "${round}_${i}" &
    pids+=($!)
  done
  for pid in "${pids[@]}"; do
    if ! wait "$pid"; then
      FAILED=$((FAILED + 1))
    fi
  done
done

if [[ "$FAILED" -gt 0 ]]; then
  echo "Load smoke FAILED: $FAILED worker(s)"
  exit 1
fi
echo "Load smoke PASSED"
