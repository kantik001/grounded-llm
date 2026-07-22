#!/usr/bin/env bash
# Postgres dump → restore smoke (CI / local). Verifies migrations + pg_dump/pg_restore round-trip.
#
# Usage:
#   PGHOST=127.0.0.1 PGUSER=grounded PGPASSWORD=grounded PGDATABASE=grounded \
#     bash scripts/backup_postgres_smoke.sh
#
# Or: DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=disable bash scripts/backup_postgres_smoke.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DUMP_FILE="${DUMP_FILE:-/tmp/grounded-backup-smoke.dump}"
RESTORE_DB="${RESTORE_DB:-grounded_restore_smoke}"
MARKER_TG_ID="${MARKER_TG_ID:-910000001}"

if ! command -v psql >/dev/null 2>&1 || ! command -v pg_dump >/dev/null 2>&1 || ! command -v pg_restore >/dev/null 2>&1; then
  echo "Need psql, pg_dump, and pg_restore (postgresql-client) on PATH"
  exit 1
fi

# Prefer explicit PG* ; optionally derive from DATABASE_URL
if [[ -n "${DATABASE_URL:-}" ]]; then
  # postgres://user:pass@host:port/db?sslmode=disable
  _url="${DATABASE_URL#postgres://}"
  _url="${_url#postgresql://}"
  _creds="${_url%%@*}"
  _rest="${_url#*@}"
  export PGUSER="${PGUSER:-${_creds%%:*}}"
  export PGPASSWORD="${PGPASSWORD:-${_creds#*:}}"
  _hostport="${_rest%%/*}"
  _dbq="${_rest#*/}"
  export PGHOST="${PGHOST:-${_hostport%%:*}}"
  if [[ "$_hostport" == *:* ]]; then
    export PGPORT="${PGPORT:-${_hostport##*:}}"
  else
    export PGPORT="${PGPORT:-5432}"
  fi
  export PGDATABASE="${PGDATABASE:-${_dbq%%\?*}}"
fi

export PGHOST="${PGHOST:-127.0.0.1}"
export PGPORT="${PGPORT:-5432}"
export PGUSER="${PGUSER:-grounded}"
export PGPASSWORD="${PGPASSWORD:-grounded}"
export PGDATABASE="${PGDATABASE:-grounded}"
export PGSSLMODE="${PGSSLMODE:-disable}"

echo "Backup smoke: host=${PGHOST}:${PGPORT} db=${PGDATABASE} user=${PGUSER}"

echo "Apply migrations"
mapfile -t MIGRATIONS < <(ls -1 "$ROOT"/migrations/*.sql | sort)
for f in "${MIGRATIONS[@]}"; do
  echo "  $(basename "$f")"
  psql -v ON_ERROR_STOP=1 -f "$f" >/dev/null
done

echo "Seed marker user telegram_id=${MARKER_TG_ID}"
psql -v ON_ERROR_STOP=1 <<SQL
INSERT INTO users (telegram_id, username, first_name, last_name)
VALUES (${MARKER_TG_ID}, 'backup_smoke', 'Backup', 'Smoke')
ON CONFLICT (telegram_id) DO UPDATE
SET username = EXCLUDED.username, updated_at = NOW();
SQL

COUNT=$(psql -tAc "SELECT count(*) FROM users WHERE telegram_id = ${MARKER_TG_ID}")
COUNT=$(echo "$COUNT" | tr -d '[:space:]')
if [[ "$COUNT" != "1" ]]; then
  echo "FAIL: seed user missing (count=${COUNT})"
  exit 1
fi

echo "pg_dump → ${DUMP_FILE}"
pg_dump -Fc -f "$DUMP_FILE"

echo "Recreate ${RESTORE_DB}"
psql -d postgres -v ON_ERROR_STOP=1 -c \
  "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '${RESTORE_DB}' AND pid <> pg_backend_pid();" \
  >/dev/null 2>&1 || true
psql -d postgres -v ON_ERROR_STOP=1 -c "DROP DATABASE IF EXISTS ${RESTORE_DB};"
psql -d postgres -v ON_ERROR_STOP=1 -c "CREATE DATABASE ${RESTORE_DB};"

echo "pg_restore into ${RESTORE_DB}"
# Fresh DB: restore without --clean (no objects to drop yet)
pg_restore -d "$RESTORE_DB" --no-owner --no-acl "$DUMP_FILE"

RESTORED=$(psql -d "$RESTORE_DB" -tAc "SELECT count(*) FROM users WHERE telegram_id = ${MARKER_TG_ID}")
RESTORED=$(echo "$RESTORED" | tr -d '[:space:]')
if [[ "$RESTORED" != "1" ]]; then
  echo "FAIL: restored user missing (count=${RESTORED})"
  exit 1
fi

psql -d postgres -v ON_ERROR_STOP=1 -c "DROP DATABASE IF EXISTS ${RESTORE_DB};" >/dev/null
rm -f "$DUMP_FILE"

echo "Backup smoke PASSED"
