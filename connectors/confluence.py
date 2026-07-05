"""Confluence Cloud connector via REST API."""

from __future__ import annotations

import os
import re
from pathlib import Path

import requests

from connectors._html import html_to_text
from connectors.base import Connector, SyncResult
from connectors.file_sync import SUPPORTED_SUFFIXES


def _safe_filename(title: str) -> str:
    cleaned = re.sub(r"[^\w.\- ]+", "_", title).strip() or "page"
    return cleaned[:160]


class ConfluenceConnector(Connector):
    """
    Export Confluence pages (HTML → txt) and downloadable attachments.

    Required env:
      CONFLUENCE_BASE_URL   e.g. https://your.atlassian.net/wiki
      CONFLUENCE_EMAIL
      CONFLUENCE_API_TOKEN
      CONFLUENCE_SPACE_KEY
    """

    name = "confluence"

    def __init__(
        self,
        *,
        base_url: str | None = None,
        email: str | None = None,
        api_token: str | None = None,
        space_key: str | None = None,
    ) -> None:
        self.base_url = (base_url or os.environ.get("CONFLUENCE_BASE_URL") or "").rstrip("/")
        self.email = (email or os.environ.get("CONFLUENCE_EMAIL") or "").strip()
        self.api_token = (api_token or os.environ.get("CONFLUENCE_API_TOKEN") or "").strip()
        self.space_key = (space_key or os.environ.get("CONFLUENCE_SPACE_KEY") or "").strip()

        missing = [
            n
            for n, v in (
                ("CONFLUENCE_BASE_URL", self.base_url),
                ("CONFLUENCE_EMAIL", self.email),
                ("CONFLUENCE_API_TOKEN", self.api_token),
                ("CONFLUENCE_SPACE_KEY", self.space_key),
            )
            if not v
        ]
        if missing:
            raise ValueError(f"Missing Confluence config: {', '.join(missing)}")

    def _session(self) -> requests.Session:
        sess = requests.Session()
        sess.auth = (self.email, self.api_token)
        sess.headers.update({"Accept": "application/json"})
        return sess

    def _list_pages(self, sess: requests.Session) -> list[dict]:
        url = f"{self.base_url}/rest/api/content"
        params = {
            "spaceKey": self.space_key,
            "type": "page",
            "expand": "body.storage,children.page",
            "limit": 100,
        }
        pages: list[dict] = []
        while url:
            resp = sess.get(url, params=params if "rest/api/content" in url else None, timeout=60)
            resp.raise_for_status()
            data = resp.json()
            pages.extend(data.get("results") or [])
            next_link = (data.get("_links") or {}).get("next")
            url = f"{self.base_url}{next_link}" if next_link else ""
            params = None
        return pages

    def _list_attachments(self, sess: requests.Session, page_id: str) -> list[dict]:
        url = f"{self.base_url}/rest/api/content/{page_id}/child/attachment"
        resp = sess.get(url, timeout=60)
        resp.raise_for_status()
        return resp.json().get("results") or []

    def sync(self, target_dir: Path, *, dry_run: bool = False) -> SyncResult:
        result = SyncResult(connector=self.name)
        target_dir.mkdir(parents=True, exist_ok=True)
        sess = self._session()

        for page in self._list_pages(sess):
            title = page.get("title") or "page"
            page_id = page.get("id")
            body = ((page.get("body") or {}).get("storage") or {}).get("value") or ""
            text = html_to_text(body)
            if text:
                fname = f"{_safe_filename(title)}.txt"
                if dry_run:
                    result.files_copied += 1
                else:
                    (target_dir / fname).write_text(text, encoding="utf-8")
                    result.files_copied += 1

            for att in self._list_attachments(sess, page_id):
                att_title = att.get("title") or "attachment"
                suffix = Path(att_title).suffix.lower()
                if suffix not in SUPPORTED_SUFFIXES:
                    result.files_skipped += 1
                    continue
                download = ((att.get("_links") or {}).get("download") or "").strip()
                if not download:
                    result.files_skipped += 1
                    continue
                if dry_run:
                    result.files_copied += 1
                    continue
                resp = sess.get(f"{self.base_url}{download}", timeout=120)
                resp.raise_for_status()
                out = target_dir / _safe_filename(att_title)
                out.write_bytes(resp.content)
                result.files_copied += 1

        return result
