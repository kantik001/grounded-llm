"""Named connector registry for CLI and programmatic use."""

from __future__ import annotations

from collections.abc import Callable
from typing import Type

from connectors.base import Connector
from connectors.confluence import ConfluenceConnector
from connectors.confluence_export import ConfluenceExportConnector
from connectors.google_drive import GoogleDriveConnector
from connectors.google_drive_export import GoogleDriveExportConnector
from connectors.local_folder import LocalFolderConnector
from connectors.sharepoint import SharePointGraphConnector
from connectors.sharepoint_export import SharePointExportConnector

# Folder / offline-export connectors: constructed as Class(source_path).
FOLDER_CONNECTORS: dict[str, Type[Connector]] = {
    "local_folder": LocalFolderConnector,
    "sharepoint_export": SharePointExportConnector,
    "google_drive_export": GoogleDriveExportConnector,
    "confluence_export": ConfluenceExportConnector,
}

# Live API connectors: factory(source) — source is optional SharePoint subfolder.
API_CONNECTORS: dict[str, Callable[[str], Connector]] = {
    "sharepoint": lambda source: SharePointGraphConnector(folder_path=source or ""),
    "google_drive": lambda _source: GoogleDriveConnector(),
    "confluence": lambda _source: ConfluenceConnector(),
}

CONNECTOR_NAMES: list[str] = sorted({*FOLDER_CONNECTORS, *API_CONNECTORS})


def requires_source(name: str) -> bool:
    return name in FOLDER_CONNECTORS


def build_connector(name: str, source: str = "") -> Connector:
    if name in FOLDER_CONNECTORS:
        return FOLDER_CONNECTORS[name](source)
    if name in API_CONNECTORS:
        return API_CONNECTORS[name](source)
    known = ", ".join(CONNECTOR_NAMES)
    raise ValueError(f"Unknown connector: {name}. Choose one of: {known}")
