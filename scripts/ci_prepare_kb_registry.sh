#!/usr/bin/env bash
# Apply migrations and backfill kb_documents from data/ (CI / local eval gate).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "DATABASE_URL is required for KB registry seed"
  exit 1
fi

export DATA_DIR="${DATA_DIR:-$ROOT/data}"
export KB_BLOB_BACKEND="${KB_BLOB_BACKEND:-local}"
export KB_BLOB_DIR="${KB_BLOB_DIR:-/tmp/grounded-kb-blobs}"
mkdir -p "$KB_BLOB_DIR"

PSQL_URL="${DATABASE_URL/postgresql+psycopg/postgresql}"

echo "==> Apply Postgres migrations"
mapfile -t MIGRATIONS < <(ls -1 "$ROOT"/migrations/*.sql | sort)
for f in "${MIGRATIONS[@]}"; do
  echo "  $(basename "$f")"
  psql "$PSQL_URL" -v ON_ERROR_STOP=1 -f "$f" >/dev/null
done

echo "==> Backfill KB registry from ${DATA_DIR}"
python scripts/backfill_kb_registry.py
