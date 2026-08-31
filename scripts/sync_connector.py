#!/usr/bin/env python3
"""Sync external documents into KB registry via a connector."""

from __future__ import annotations

import argparse
import os
import shutil
import sys

_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
_SCRIPTS = os.path.dirname(__file__)
for _p in (_ROOT, _SCRIPTS):
    if _p not in sys.path:
        sys.path.insert(0, _p)

from connectors.registry import (  # noqa: E402
    CONNECTOR_NAMES,
    build_connector,
    requires_source,
)

from pack_installer import connector_staging_dir  # noqa: E402


def main() -> int:
    parser = argparse.ArgumentParser(description="Sync KB documents via connector")
    parser.add_argument("connector", choices=CONNECTOR_NAMES, help="Connector name")
    parser.add_argument(
        "--source",
        default="",
        help="Folder path (export connectors) or SharePoint subfolder",
    )
    parser.add_argument("--tenant", default="default")
    parser.add_argument("--domain", required=True, help="Domain id")
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args()

    if requires_source(args.connector) and not args.source:
        print("--source is required for folder/export connectors", file=sys.stderr)
        return 1

    target = connector_staging_dir(args.tenant, args.domain)
    try:
        conn = build_connector(args.connector, args.source)
    except (FileNotFoundError, ValueError, RuntimeError) as exc:
        print(f"Error: {exc}", file=sys.stderr)
        return 1

    result = conn.sync(target, dry_run=args.dry_run)
    print(
        f"{result.connector}: copied={result.files_copied} skipped={result.files_skipped} "
        f"staging={target}"
    )
    for err in result.errors:
        print(f"  error: {err}", file=sys.stderr)
    if args.dry_run:
        print("(dry run — no files written)")
        return 0 if result.ok else 1

    if not result.ok:
        return 1

    from connectors.registry_sync import register_synced_tree

    ids = register_synced_tree(
        tenant_id=args.tenant,
        domain_id=args.domain,
        target_dir=target,
        source=args.connector,
    )
    print(f"registered {len(ids)} document(s) into KB registry")
    shutil.rmtree(target, ignore_errors=True)

    from rag.kb.outbox import auto_ingest_enabled, flush_outbox

    if auto_ingest_enabled():
        result = flush_outbox(tenant_id=args.tenant, domain_id=args.domain, sync=False)
        if result.error:
            print(f"auto-ingest flush failed: {result.error}", file=sys.stderr)
            return 1
        if result.flushed:
            msg = f"auto-ingest queued job_id={result.job_id} flushed={result.flushed}"
            if result.already_running:
                msg += " (already running)"
            print(msg)

    return 0


if __name__ == "__main__":
    sys.exit(main())
