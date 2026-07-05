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

from connectors.base import Connector  # noqa: E402
from connectors.confluence_export import ConfluenceExportConnector  # noqa: E402
from connectors.google_drive_export import GoogleDriveExportConnector  # noqa: E402
from connectors.local_folder import LocalFolderConnector  # noqa: E402
from connectors.sharepoint import SharePointGraphConnector  # noqa: E402
from connectors.sharepoint_export import SharePointExportConnector  # noqa: E402

from pack_installer import data_target_dir  # noqa: E402

CONNECTORS = {
    "local_folder": LocalFolderConnector,
    "sharepoint_export": SharePointExportConnector,
    "google_drive_export": GoogleDriveExportConnector,
    "confluence_export": ConfluenceExportConnector,
    "sharepoint": SharePointGraphConnector,
}


def build_connector(name: str, source: str) -> Connector:
    if name in ("local_folder", "sharepoint_export", "google_drive_export", "confluence_export"):
        return CONNECTORS[name](source)
    if name == "sharepoint":
        return SharePointGraphConnector(folder_path=source or "")
    raise ValueError(f"Unknown connector: {name}")


def main() -> int:
    parser = argparse.ArgumentParser(description="Sync KB documents via connector")
    parser.add_argument("connector", choices=sorted(CONNECTORS), help="Connector name")
    parser.add_argument(
        "--source",
        default="",
        help="Source path (folder connectors) or SharePoint subfolder path",
    )
    parser.add_argument("--tenant", default="default")
    parser.add_argument("--domain", required=True, help="Domain id")
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args()

    if args.connector != "sharepoint" and not args.source:
        print("--source is required for folder connectors", file=sys.stderr)
        return 1

    target = data_target_dir(args.tenant, args.domain)
    try:
        conn = build_connector(args.connector, args.source)
    except (FileNotFoundError, ValueError) as exc:
        print(f"Error: {exc}", file=sys.stderr)
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
