"""Connector interface for knowledge-base ingest."""

from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from pathlib import Path


@dataclass
class SyncResult:
    connector: str
    files_copied: int = 0
    files_skipped: int = 0
    errors: list[str] = field(default_factory=list)

    @property
    def ok(self) -> bool:
        return not self.errors


class Connector(ABC):
    """Copy or download documents into data/{tenant}/{domain}/."""

    name: str = "base"

    @abstractmethod
    def sync(self, target_dir: Path, *, dry_run: bool = False) -> SyncResult:
        """Populate target_dir with supported document files."""
