#!/usr/bin/env bash
# LLM E2E smoke: real LLM API + mocked RAG (no Python/Chroma required).
# Requires: running Go server, Postgres, LLM_API_KEY, TELEGRAM_AUTH_DISABLED=true, RAG_MOCK=true
set -euo pipefail

BASE_URL="${1:-http://127.0.0.1:8080}"

if [[ -z "${LLM_API_KEY:-}" ]]; then
  echo "Skip LLM E2E: LLM_API_KEY not set"
  exit 0
fi

echo "LLM E2E smoke: $BASE_URL (RAG_MOCK expected on server)"

SESSION=$(curl -sS -X POST "${BASE_URL}/api/session" \
  -H "Content-Type: application/json" \
  -d '{"domain_id":"default"}' | python -c "import sys,json; print(json.load(sys.stdin).get('session_id',''))")

if [[ -z "$SESSION" ]]; then
  echo "FAIL: no session_id"
  exit 1
fi

BODY=$(curl -sS -X POST "${BASE_URL}/api/message" \
  -H "Content-Type: application/json" \
  -d "{\"session_id\":\"${SESSION}\",\"domain_id\":\"default\",\"text\":\"How many paid vacation days do employees get?\"}")

echo "$BODY" | python -c "
import json, sys
data = json.load(sys.stdin)
if not data.get('success'):
    raise SystemExit('FAIL: success=false')
msgs = data.get('messages') or []
assistant = next((m for m in reversed(msgs) if m.get('role')=='assistant'), None)
if not assistant:
    raise SystemExit('FAIL: no assistant message')
text = assistant.get('content') or ''
if '28' not in text:
    raise SystemExit(f'FAIL: expected 28 in answer, got: {text[:120]}')
cites = assistant.get('citations') or []
if not cites:
    raise SystemExit('FAIL: expected citations')
print('LLM E2E PASSED')
print('Answer preview:', text[:160].replace(chr(10), ' '))
"
