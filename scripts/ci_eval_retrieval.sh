#!/usr/bin/env bash
# CI / local: reindex Chroma, start Python RAG, run retrieval eval suites (no LLM).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

SKIP_REINDEX=false
for arg in "$@"; do
  case "$arg" in
    --skip-reindex) SKIP_REINDEX=true ;;
  esac
done

export DOMAINS_CONFIG_PATH="${DOMAINS_CONFIG_PATH:-$ROOT/config/domains.json}"
export LOCALES_ROOT="${LOCALES_ROOT:-$ROOT/config/locales}"
export DEFAULT_LOCALE="${DEFAULT_LOCALE:-en}"
export FORCE_RAG_REINDEX="${FORCE_RAG_REINDEX:-true}"
export PYTHON_SERVICE_PORT="${PYTHON_SERVICE_PORT:-5000}"
RAG_URL="http://127.0.0.1:${PYTHON_SERVICE_PORT}/rag/context"

if [ "$SKIP_REINDEX" = false ]; then
  echo "==> Reindexing Chroma (FORCE_RAG_REINDEX=${FORCE_RAG_REINDEX})"
  python scripts/reindex_rag.py
else
  echo "==> Skipping reindex (--skip-reindex); using existing chroma_db"
  export FORCE_RAG_REINDEX=false
fi

echo "==> Starting Python RAG on :${PYTHON_SERVICE_PORT}"
python api/app.py &
APP_PID=$!
cleanup() {
  kill "$APP_PID" 2>/dev/null || true
  wait "$APP_PID" 2>/dev/null || true
}
trap cleanup EXIT

for i in $(seq 1 90); do
  if curl -sf "http://127.0.0.1:${PYTHON_SERVICE_PORT}/health" >/dev/null; then
    echo "Python RAG ready"
    break
  fi
  if [ "$i" -eq 90 ]; then
    echo "Python RAG failed to start within 180s"
    exit 1
  fi
  sleep 2
done

echo "==> Running retrieval eval (all suites)"
python scripts/run_rag_eval.py --suite all --rag-url "$RAG_URL"
echo "==> Retrieval eval passed"
