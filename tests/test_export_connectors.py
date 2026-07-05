"""Tests for export connectors."""

import pytest
from connectors.confluence_export import ConfluenceExportConnector
from connectors.google_drive_export import GoogleDriveExportConnector
from connectors.sharepoint_export import SharePointExportConnector


@pytest.mark.parametrize(
    "connector_cls",
    [SharePointExportConnector, GoogleDriveExportConnector, ConfluenceExportConnector],
)
def test_export_connectors_dry_run(tmp_path, connector_cls):
    src = tmp_path / "src"
    src.mkdir()
    (src / "doc.txt").write_text("policy text", encoding="utf-8")

    dest = tmp_path / "dest"
    conn = connector_cls(src)
    result = conn.sync(dest, dry_run=True)
    assert result.ok
    assert result.files_copied == 1
