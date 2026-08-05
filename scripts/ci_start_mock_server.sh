#!/usr/bin/env bash
# Build Go server with LLM/RAG mocks and wait until /health is ready.
# Env (required/typical): DATABASE_URL, DOMAINS_CONFIG_PATH, LOCALES_ROOT,
# MIGRATIONS_DIR, DATA_DIR, SERVER_PORT (default 8080).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

export TELEGRAM_AUTH_DISABLED="${TELEGRAM_AUTH_DISABLED:-true}"
export LLM_MOCK="${LLM_MOCK:-true}"
export RAG_MOCK="${RAG_MOCK:-true}"
export PYTHON_RAG_URL="${PYTHON_RAG_URL:-http://127.0.0.1:59999/rag/context}"
export PYTHON_BASE_URL="${PYTHON_BASE_URL:-http://127.0.0.1:59999}"
export UPLOAD_DIR="${UPLOAD_DIR:-/tmp/grounded-uploads}"
export SERVER_PORT="${SERVER_PORT:-8080}"
export MIGRATIONS_DIR="${MIGRATIONS_DIR:-$ROOT/migrations}"
export DATA_DIR="${DATA_DIR:-$ROOT/data}"
export DOMAINS_CONFIG_PATH="${DOMAINS_CONFIG_PATH:-$ROOT/config/domains.json}"
export LOCALES_ROOT="${LOCALES_ROOT:-$ROOT/config/locales}"

mkdir -p "$UPLOAD_DIR"
cd server
go mod tidy
go build -o ../server-bin .
cd ..

./server-bin &
echo $! > server.pid

for i in $(seq 1 60); do
  if curl -sf "http://127.0.0.1:${SERVER_PORT}/health" >/dev/null; then
    echo "Server ready on :${SERVER_PORT}"
    exit 0
  fi
  sleep 1
done

echo "Server failed to start within 60s"
exit 1
