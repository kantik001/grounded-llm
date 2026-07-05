"""SharePoint Online connector via Microsoft Graph API (app-only)."""

from __future__ import annotations

import os
from pathlib import Path
from typing import Any
from urllib.parse import quote

import requests

from connectors.base import Connector, SyncResult
from connectors.file_sync import SUPPORTED_SUFFIXES

GRAPH_ROOT = "https://graph.microsoft.com/v1.0"


class SharePointGraphConnector(Connector):
    """
    Download documents from a SharePoint / OneDrive drive via Graph.

    Required env:
      SHAREPOINT_TENANT_ID
      SHAREPOINT_CLIENT_ID
      SHAREPOINT_CLIENT_SECRET
      SHAREPOINT_DRIVE_ID   (drive id from Graph)
    """

    name = "sharepoint"

    def __init__(
        self,
        *,
        tenant_id: str | None = None,
        client_id: str | None = None,
        client_secret: str | None = None,
        drive_id: str | None = None,
        folder_path: str = "",
    ) -> None:
        self.tenant_id = (tenant_id or os.environ.get("SHAREPOINT_TENANT_ID") or "").strip()
        self.client_id = (client_id or os.environ.get("SHAREPOINT_CLIENT_ID") or "").strip()
        self.client_secret = (
            client_secret or os.environ.get("SHAREPOINT_CLIENT_SECRET") or ""
        ).strip()
        self.drive_id = (drive_id or os.environ.get("SHAREPOINT_DRIVE_ID") or "").strip()
        self.folder_path = (folder_path or os.environ.get("SHAREPOINT_FOLDER_PATH") or "").strip("/")

        missing = [
            name
            for name, val in (
                ("SHAREPOINT_TENANT_ID", self.tenant_id),
                ("SHAREPOINT_CLIENT_ID", self.client_id),
                ("SHAREPOINT_CLIENT_SECRET", self.client_secret),
                ("SHAREPOINT_DRIVE_ID", self.drive_id),
            )
            if not val
        ]
        if missing:
            raise ValueError(f"Missing SharePoint Graph config: {', '.join(missing)}")

    def _token(self) -> str:
        url = f"https://login.microsoftonline.com/{self.tenant_id}/oauth2/v2.0/token"
        resp = requests.post(
            url,
            data={
                "client_id": self.client_id,
                "client_secret": self.client_secret,
                "scope": "https://graph.microsoft.com/.default",
                "grant_type": "client_credentials",
            },
            timeout=60,
        )
        resp.raise_for_status()
        return resp.json()["access_token"]

    def _headers(self, token: str) -> dict[str, str]:
        return {"Authorization": f"Bearer {token}"}

    def _list_children(self, token: str, item_path: str) -> list[dict[str, Any]]:
        base = f"{GRAPH_ROOT}/drives/{self.drive_id}/root"
        if item_path:
            base = f"{base}:/{quote(item_path)}:"
        url = f"{base}/children"
        items: list[dict[str, Any]] = []
        while url:
            resp = requests.get(url, headers=self._headers(token), timeout=60)
            resp.raise_for_status()
            payload = resp.json()
            items.extend(payload.get("value") or [])
            url = payload.get("@odata.nextLink")
        return items

    def _download_file(self, token: str, item: dict[str, Any], dest: Path) -> None:
        url = item.get("@microsoft.graph.downloadUrl")
        if not url:
            item_id = item["id"]
            url = f"{GRAPH_ROOT}/drives/{self.drive_id}/items/{item_id}/content"
        resp = requests.get(url, headers=self._headers(token), timeout=120)
        resp.raise_for_status()
        dest.parent.mkdir(parents=True, exist_ok=True)
        dest.write_bytes(resp.content)

    def _walk(
        self,
        token: str,
        rel_prefix: str,
        target_dir: Path,
        result: SyncResult,
        *,
        dry_run: bool,
    ) -> None:
        path = "/".join(p for p in (self.folder_path, rel_prefix) if p)
        for item in self._list_children(token, path):
            name = item.get("name") or "unknown"
            rel = f"{rel_prefix}/{name}".strip("/")
            if item.get("folder"):
                self._walk(token, rel, target_dir, result, dry_run=dry_run)
                continue
            suffix = Path(name).suffix.lower()
            if suffix not in SUPPORTED_SUFFIXES:
                result.files_skipped += 1
                continue
            dest = target_dir / rel
            if dry_run:
                result.files_copied += 1
                continue
            try:
                self._download_file(token, item, dest)
                result.files_copied += 1
            except OSError as exc:
                result.errors.append(f"{rel}: {exc}")

    def sync(self, target_dir: Path, *, dry_run: bool = False) -> SyncResult:
        result = SyncResult(connector=self.name)
        target_dir.mkdir(parents=True, exist_ok=True)
        token = self._token()
        self._walk(token, "", target_dir, result, dry_run=dry_run)
        return result
