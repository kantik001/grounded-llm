"""Ingest connectors — sync external sources into the knowledge base data directory.

Registry: ``connectors.registry.build_connector``. CLI: ``scripts/sync_connector.py``.
Optional Google Drive deps: ``pip install -r connectors/requirements.txt``.
"""

from connectors.base import Connector, SyncResult
from connectors.registry import CONNECTOR_NAMES, build_connector, requires_source

__all__ = [
    "CONNECTOR_NAMES",
    "Connector",
    "SyncResult",
    "build_connector",
    "requires_source",
]
