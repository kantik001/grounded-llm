#!/usr/bin/env python3
"""Garbage-collect retired index runs (Chroma persist dirs + sparse BM25 pickles)."""

from __future__ import annotations

import argparse
import os
import shutil
import sys

_PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
if _PROJECT_ROOT not in sys.path:
    sys.path.insert(0, _PROJECT_ROOT)


def _sparse_base_dir() -> str:
    return (os.environ.get("SPARSE_INDEX_DIR") or os.path.join(_PROJECT_ROOT, "sparse_index")).strip()


def _delete_chroma_run(tenant_id: str, domain_id: str, run_id: str, *, dry_run: bool) -> str | None:
    from rag.kb.index_collections import chroma_run_dir
    from rag.vector_backend.chroma_backend import _persist_dir

    path = chroma_run_dir(_persist_dir(), tenant_id, domain_id, run_id)
    if not os.path.isdir(path):
        return None
    if dry_run:
        print(f"[dry-run] would remove chroma run dir: {path}")
        return path
    shutil.rmtree(path, ignore_errors=True)
    print(f"removed chroma run dir: {path}")
    return path


def _delete_sparse_run(tenant_id: str, domain_id: str, run_id: str, *, dry_run: bool) -> str | None:
    from rag.kb.index_collections import sparse_run_dir

    path = sparse_run_dir(_sparse_base_dir(), tenant_id, domain_id, run_id)
    if not os.path.isdir(path):
        return None
    if dry_run:
        print(f"[dry-run] would remove sparse run dir: {path}")
        return path
    shutil.rmtree(path, ignore_errors=True)
    print(f"removed sparse run dir: {path}")
    return path


def main() -> int:
    from rag.kb.index_collections import run_suffix
    from rag.kb.index_runs import list_retired_run_ids

    parser = argparse.ArgumentParser(description="GC retired index runs for a tenant/domain scope")
    parser.add_argument("--tenant", default="default", help="tenant id")
    parser.add_argument("--domain", default="default", help="domain id")
    parser.add_argument("--keep-last", type=int, default=1, help="retired runs to keep (newest)")
    parser.add_argument("--dry-run", action="store_true", help="print actions without deleting")
    args = parser.parse_args()

    run_ids = list_retired_run_ids(args.tenant, args.domain, keep_last=args.keep_last)
    if not run_ids:
        print(f"no retired runs to GC for {args.tenant}/{args.domain}")
        return 0

    print(f"GC {len(run_ids)} retired run(s) for {args.tenant}/{args.domain}")
    for run_id in run_ids:
        suffix = run_suffix(run_id)
        print(f"  run {run_id} ({suffix})")
        _delete_chroma_run(args.tenant, args.domain, run_id, dry_run=args.dry_run)
        _delete_sparse_run(args.tenant, args.domain, run_id, dry_run=args.dry_run)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
