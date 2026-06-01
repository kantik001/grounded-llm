#!/usr/bin/env bash
# Create a new knowledge domain (domain pack scaffold).
set -euo pipefail

DOMAIN_ID="${1:-}"
TENANT_ID="${2:-default}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

if [[ -z "$DOMAIN_ID" ]]; then
  echo "Usage: $0 <domain_id> [tenant_id]"
  exit 1
fi

if [[ ! "$DOMAIN_ID" =~ ^[a-z][a-z0-9_]*$ ]]; then
  echo "domain_id must be lowercase slug: letters, digits, underscore"
  exit 1
fi

DATA_DIR="$ROOT/data/$TENANT_ID/$DOMAIN_ID"
mkdir -p "$DATA_DIR"

echo "Created $DATA_DIR"
echo "Next:"
echo "  1. Add entry to config/domains.json"
echo "  2. Put .txt/.pdf/.docx files in $DATA_DIR"
echo "  3. Update config/prompts.json, few_shot.json, onboarding.json"
echo "  4. python scripts/reindex_rag.py"
