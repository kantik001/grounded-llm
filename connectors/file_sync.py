"""Shared file-tree sync helpers for folder-based connectors."""

from __future__ import annotations

import shutil
from pathlib import Path

from connectors.base import SyncResult

SUPPORTED_SUFFIXES = {".txt", ".pdf", ".docx"}


def sync_file_tree(
    source_dir: Path,
    target_dir: Path,
    *,
    connector_name: str,
    dry_run: bool = False,
    include_suffixes: set[str] | None = None,
) -> SyncResult:
    source_dir = source_dir.resolve()
    if not source_dir.is_dir():
        raise FileNotFoundError(f"Source directory not found: {source_dir}")

    suffixes = include_suffixes or SUPPORTED_SUFFIXES
    result = SyncResult(connector=connector_name)
    target_dir.mkdir(parents=True, exist_ok=True)

    for src in sorted(source_dir.rglob("*")):
        if not src.is_file():
            continue
        if src.suffix.lower() not in suffixes:
            result.files_skipped += 1
            continue
        rel = src.relative_to(source_dir)
        dest = target_dir / rel
        if dry_run:
            result.files_copied += 1
            continue
        dest.parent.mkdir(parents=True, exist_ok=True)
        try:
            shutil.copy2(src, dest)
            result.files_copied += 1
        except OSError as exc:
            result.errors.append(f"{rel}: {exc}")

    return result
