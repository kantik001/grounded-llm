#!/usr/bin/env python3
"""Flush kb_ingest_outbox pending rows → Go POST /admin/ingest."""

from __future__ import annotations

import argparse
import os
import sys

_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
if _ROOT not in sys.path:
    sys.path.insert(0, _ROOT)

from rag.kb.outbox import flush_outbox  # noqa: E402


def main() -> int:
    parser = argparse.ArgumentParser(description="Flush KB ingest outbox → admin ingest API")
    parser.add_argument("--tenant", default="default")
    parser.add_argument("--domain", required=True)
    parser.add_argument("--sync", action="store_true", help="Run ingest synchronously in Python API")
    args = parser.parse_args()

    result = flush_outbox(tenant_id=args.tenant, domain_id=args.domain, sync=args.sync)
    if result.error:
        print(f"FAIL: {result.error}", file=sys.stderr)
        return 1
    if result.flushed == 0:
        print("no pending outbox rows")
        return 0
    print(
        f"flushed={result.flushed} job_id={result.job_id} "
        f"already_running={result.already_running}"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
