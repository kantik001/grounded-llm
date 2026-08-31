#!/usr/bin/env python3
"""Backfill kb_documents registry from legacy data/{tenant}/{domain}/ files."""

from __future__ import annotations

import argparse
import os
import sys

# Repo root on path
_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
if _ROOT not in sys.path:
    sys.path.insert(0, _ROOT)


def main() -> int:
    parser = argparse.ArgumentParser(description="Backfill KB document registry from data/ tree")
    parser.add_argument("--tenant", default=os.environ.get("DEFAULT_TENANT_ID", "default"))
    parser.add_argument("--domain", default="")
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args()

    from rag.document_loaders import is_supported_filename
    from rag.kb.documents import upsert_document
    from rag.kb_discovery import discover_kb_directories

    total = 0
    for tenant_id, domain_id, domain_dir in discover_kb_directories():
        if args.tenant and tenant_id != args.tenant:
            continue
        if args.domain and domain_id != args.domain:
            continue
        for name in sorted(os.listdir(domain_dir)):
            path = os.path.join(domain_dir, name)
            if not os.path.isfile(path) or not is_supported_filename(name):
                continue
            data = open(path, "rb").read()
            if args.dry_run:
                print(f"would register {tenant_id}/{domain_id}/{name} ({len(data)} bytes)")
            else:
                doc = upsert_document(
                    tenant_id=tenant_id,
                    domain_id=domain_id,
                    logical_key=name,
                    data=data,
                    source="backfill",
                    source_ref={"path": path},
                )
                print(f"registered {doc.logical_key} -> {doc.id} v{doc.current_version}")
            total += 1

    print(f"done: {total} file(s)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
