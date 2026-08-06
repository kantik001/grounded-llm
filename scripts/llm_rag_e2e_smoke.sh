#!/usr/bin/env bash
# Real RAG + real LLM E2E: expects indexed KB and RAG_MOCK=false on the Go server.
# Soft assertions: success, non-empty answer, citations present; vacation day "28"
# is preferred but not required (LLM wording may vary).
set -euo pipefail

BASE_URL="${1:-http://127.0.0.1:8080}"

if [[ -z "${LLM_API_KEY:-}" ]]; then
  echo "Skip real RAG+LLM E2E: LLM_API_KEY not set"
  exit 0
fi

echo "Real RAG+LLM E2E: $BASE_URL"

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
    raise SystemExit('FAIL: success=false err=' + str(data.get('error')))
msgs = data.get('messages') or []
assistant = next((m for m in reversed(msgs) if m.get('role')=='assistant'), None)
if not assistant:
    raise SystemExit('FAIL: no assistant message')
text = (assistant.get('content') or '').strip()
if len(text) < 8:
    raise SystemExit('FAIL: empty/short answer')
cites = assistant.get('citations') or []
if not cites:
    raise SystemExit('FAIL: expected citations from real RAG')
# Prefer grounded vacation figure when present.
if '28' in text:
    print('Real RAG+LLM E2E PASSED (grounded 28)')
else:
    print('Real RAG+LLM E2E PASSED (citations present; answer may paraphrase)')
print('Answer preview:', text[:200].replace(chr(10), ' '))
print('Citations:', len(cites))
"
