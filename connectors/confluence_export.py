"""Sync from Confluence space export (PDF/Word/text attachments)."""

from __future__ import annotations

from pathlib import Path

from connectors.base import Connector, SyncResult
from connectors.file_sync import sync_file_tree


class ConfluenceExportConnector(Connector):
    """Confluence space export often includes PDFs and attachments in a folder tree."""

    name = "confluence_export"

    def __init__(self, source_dir: str | Path) -> None:
        self.source_dir = Path(source_dir)

    def sync(self, target_dir: Path, *, dry_run: bool = False) -> SyncResult:
        return sync_file_tree(self.source_dir, target_dir, connector_name=self.name, dry_run=dry_run)
