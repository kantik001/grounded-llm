"""Sync from a SharePoint document library export folder (offline export)."""

from __future__ import annotations

from pathlib import Path

from connectors.base import Connector, SyncResult
from connectors.file_sync import sync_file_tree


class SharePointExportConnector(Connector):
    """Use after exporting a SharePoint library to disk (or OneDrive sync folder)."""

    name = "sharepoint_export"

    def __init__(self, source_dir: str | Path) -> None:
        self.source_dir = Path(source_dir)

    def sync(self, target_dir: Path, *, dry_run: bool = False) -> SyncResult:
        return sync_file_tree(self.source_dir, target_dir, connector_name=self.name, dry_run=dry_run)
