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
from connectors.confluence import ConfluenceConnector  # noqa: E402
from connectors.confluence_export import ConfluenceExportConnector  # noqa: E402
from connectors.google_drive import GoogleDriveConnector  # noqa: E402
from connectors.google_drive_export import GoogleDriveExportConnector  # noqa: E402
from connectors.local_folder import LocalFolderConnector  # noqa: E402
from connectors.sharepoint import SharePointGraphConnector  # noqa: E402
from connectors.sharepoint_export import SharePointExportConnector  # noqa: E402

from pack_installer import data_target_dir  # noqa: E402

FOLDER_CONNECTORS = {
    "local_folder": LocalFolderConnector,
    "sharepoint_export": SharePointExportConnector,
    "google_drive_export": GoogleDriveExportConnector,
    "confluence_export": ConfluenceExportConnector,
}

API_CONNECTORS = {
    "sharepoint": lambda source: SharePointGraphConnector(folder_path=source or ""),
    "google_drive": lambda _source: GoogleDriveConnector(),
    "confluence": lambda _source: ConfluenceConnector(),
}

ALL_CONNECTORS = sorted({**FOLDER_CONNECTORS, **API_CONNECTORS})


def build_connector(name: str, source: str) -> Connector:
    if name in FOLDER_CONNECTORS:
        return FOLDER_CONNECTORS[name](source)
    if name in API_CONNECTORS:
        return API_CONNECTORS[name](source)
    raise ValueError(f"Unknown connector: {name}")


def main() -> int:
    parser = argparse.ArgumentParser(description="Sync KB documents via connector")
    parser.add_argument("connector", choices=ALL_CONNECTORS, help="Connector name")
    parser.add_argument(
        "--source",
        default="",
        help="Folder path (export connectors) or SharePoint subfolder",
    )
    parser.add_argument("--tenant", default="default")
    parser.add_argument("--domain", required=True, help="Domain id")
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args()

    needs_source = args.connector in FOLDER_CONNECTORS
    if needs_source and not args.source:
        print("--source is required for folder/export connectors", file=sys.stderr)
        return 1

    target = data_target_dir(args.tenant, args.domain)
    try:
        conn = build_connector(args.connector, args.source)
    except (FileNotFoundError, ValueError, RuntimeError) as exc:
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
