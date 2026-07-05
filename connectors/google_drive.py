"""Google Drive connector via service account (read-only)."""

from __future__ import annotations

import io
import os
import re
from pathlib import Path

from connectors.base import Connector, SyncResult
from connectors.file_sync import SUPPORTED_SUFFIXES

GOOGLE_EXPORT_MIME = {
    "application/vnd.google-apps.document": "text/plain",
    "application/vnd.google-apps.spreadsheet": "text/csv",
    "application/vnd.google-apps.presentation": "text/plain",
}


def _safe_filename(name: str) -> str:
    cleaned = re.sub(r"[^\w.\- ]+", "_", name).strip() or "document"
    return cleaned[:180]


class GoogleDriveConnector(Connector):
    """
    Sync supported files from a shared Drive folder.

    Required env:
      GOOGLE_APPLICATION_CREDENTIALS  (path to service account JSON)
      GOOGLE_DRIVE_FOLDER_ID
    Optional:
      GOOGLE_DRIVE_IMPERSONATE_USER   (domain-wide delegation subject)
    """

    name = "google_drive"

    def __init__(
        self,
        *,
        credentials_path: str | None = None,
        folder_id: str | None = None,
        impersonate_user: str | None = None,
    ) -> None:
        self.credentials_path = (
            credentials_path or os.environ.get("GOOGLE_APPLICATION_CREDENTIALS") or ""
        ).strip()
        self.folder_id = (folder_id or os.environ.get("GOOGLE_DRIVE_FOLDER_ID") or "").strip()
        self.impersonate_user = (
            impersonate_user or os.environ.get("GOOGLE_DRIVE_IMPERSONATE_USER") or ""
        ).strip()
        if not self.credentials_path or not Path(self.credentials_path).is_file():
            raise ValueError("GOOGLE_APPLICATION_CREDENTIALS must point to a service account JSON file")
        if not self.folder_id:
            raise ValueError("GOOGLE_DRIVE_FOLDER_ID is required")

    def _drive_service(self):
        try:
            from google.oauth2 import service_account
            from googleapiclient.discovery import build
        except ImportError as exc:
            raise RuntimeError(
                "Google Drive connector requires: pip install -r api/requirements-connectors.txt"
            ) from exc

        creds = service_account.Credentials.from_service_account_file(
            self.credentials_path,
            scopes=["https://www.googleapis.com/auth/drive.readonly"],
        )
        if self.impersonate_user:
            creds = creds.with_subject(self.impersonate_user)
        return build("drive", "v3", credentials=creds, cache_discovery=False)

    def _iter_files(self, service):
        page_token = None
        query = f"'{self.folder_id}' in parents and trashed = false"
        while True:
            resp = (
                service.files()
                .list(
                    q=query,
                    fields="nextPageToken, files(id,name,mimeType)",
                    pageToken=page_token,
                    pageSize=200,
                )
                .execute()
            )
            for item in resp.get("files", []):
                yield item
                if item.get("mimeType") == "application/vnd.google-apps.folder":
                    yield from self._iter_folder(service, item["id"])
            page_token = resp.get("nextPageToken")
            if not page_token:
                break

    def _iter_folder(self, service, folder_id: str):
        page_token = None
        query = f"'{folder_id}' in parents and trashed = false"
        while True:
            resp = (
                service.files()
                .list(
                    q=query,
                    fields="nextPageToken, files(id,name,mimeType)",
                    pageToken=page_token,
                    pageSize=200,
                )
                .execute()
            )
            for item in resp.get("files", []):
                yield item
                if item.get("mimeType") == "application/vnd.google-apps.folder":
                    yield from self._iter_folder(service, item["id"])
            page_token = resp.get("nextPageToken")
            if not page_token:
                break

    def _write_item(self, service, item: dict, dest: Path) -> None:
        mime = item.get("mimeType") or ""
        name = _safe_filename(item.get("name") or "file")
        file_id = item["id"]

        if mime == "application/vnd.google-apps.folder":
            return

        if mime in GOOGLE_EXPORT_MIME:
            export_mime = GOOGLE_EXPORT_MIME[mime]
            suffix = ".txt" if "text" in export_mime else ".csv"
            out = dest / f"{name}{suffix}"
            request = service.files().export_media(fileId=file_id, mimeType=export_mime)
        else:
            suffix = Path(name).suffix.lower()
            if suffix not in SUPPORTED_SUFFIXES:
                raise ValueError(f"unsupported suffix: {suffix}")
            out = dest / name
            request = service.files().get_media(fileId=file_id)

        from googleapiclient.http import MediaIoBaseDownload

        buffer = io.BytesIO()
        downloader = MediaIoBaseDownload(buffer, request)
        done = False
        while not done:
            _, done = downloader.next_chunk()
        out.parent.mkdir(parents=True, exist_ok=True)
        out.write_bytes(buffer.getvalue())

    def sync(self, target_dir: Path, *, dry_run: bool = False) -> SyncResult:
        result = SyncResult(connector=self.name)
        target_dir.mkdir(parents=True, exist_ok=True)
        service = self._drive_service()

        for item in self._iter_files(service):
            mime = item.get("mimeType") or ""
            name = item.get("name") or "file"
            if mime == "application/vnd.google-apps.folder":
                continue
            if mime in GOOGLE_EXPORT_MIME:
                pass
            elif Path(name).suffix.lower() not in SUPPORTED_SUFFIXES:
                result.files_skipped += 1
                continue
            if dry_run:
                result.files_copied += 1
                continue
            try:
                self._write_item(service, item, target_dir)
                result.files_copied += 1
            except (OSError, ValueError) as exc:
                result.errors.append(f"{name}: {exc}")

        return result
