#!/usr/bin/env python3
"""Sync external documents into KB data/ via a connector."""

from __future__ import annotations

import argparse
import os
import sys

_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
_SCRIPTS = os.path.dirname(__file__)
for _p in (_ROOT, _SCRIPTS):
    if _p not in sys.path:
        sys.path.insert(0, _p)

from connectors.local_folder import LocalFolderConnector  # noqa: E402

from pack_installer import data_target_dir  # noqa: E402


def main() -> int:
    parser = argparse.ArgumentParser(description="Sync KB documents via connector")
    parser.add_argument("connector", choices=["local_folder"], help="Connector name")
    parser.add_argument("--source", required=True, help="Source path (connector-specific)")
    parser.add_argument("--tenant", default="default")
    parser.add_argument("--domain", required=True, help="Domain id")
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args()

    target = data_target_dir(args.tenant, args.domain)
    if args.connector == "local_folder":
        conn = LocalFolderConnector(args.source)
    else:
        print(f"Unknown connector: {args.connector}", file=sys.stderr)
        return 1

    result = conn.sync(target, dry_run=args.dry_run)
    print(
        f"{result.connector}: copied={result.files_copied} skipped={result.files_skipped} "
        f"target={target}"
    )
    for err in result.errors:
        print(f"  error: {err}", file=sys.stderr)
    if args.dry_run:
        print("(dry run — no files written)")
    return 0 if result.ok else 1


if __name__ == "__main__":
    sys.exit(main())
