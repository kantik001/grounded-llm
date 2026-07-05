"""Reference connector: mirror a local folder into the KB data directory."""

from __future__ import annotations

import shutil
from pathlib import Path

from connectors.base import Connector, SyncResult

SUPPORTED_SUFFIXES = {".txt", ".pdf", ".docx"}


class LocalFolderConnector(Connector):
    name = "local_folder"

    def __init__(self, source_dir: str | Path) -> None:
        self.source_dir = Path(source_dir).resolve()
        if not self.source_dir.is_dir():
            raise FileNotFoundError(f"Source directory not found: {self.source_dir}")

    def sync(self, target_dir: Path, *, dry_run: bool = False) -> SyncResult:
        result = SyncResult(connector=self.name)
        target_dir.mkdir(parents=True, exist_ok=True)

        for src in sorted(self.source_dir.rglob("*")):
            if not src.is_file():
                continue
            if src.suffix.lower() not in SUPPORTED_SUFFIXES:
                result.files_skipped += 1
                continue
            rel = src.relative_to(self.source_dir)
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
