"""Tests for SharePoint Graph connector (mocked HTTP)."""

from unittest.mock import MagicMock, patch

import pytest
from connectors.sharepoint import SharePointGraphConnector


@pytest.fixture
def graph_env(monkeypatch):
    monkeypatch.setenv("SHAREPOINT_TENANT_ID", "tenant")
    monkeypatch.setenv("SHAREPOINT_CLIENT_ID", "client")
    monkeypatch.setenv("SHAREPOINT_CLIENT_SECRET", "secret")
    monkeypatch.setenv("SHAREPOINT_DRIVE_ID", "drive123")


def test_sharepoint_graph_missing_config():
    with pytest.raises(ValueError, match="Missing SharePoint Graph config"):
        SharePointGraphConnector(
            tenant_id="",
            client_id="",
            client_secret="",
            drive_id="",
        )


@patch("connectors.sharepoint.requests.post")
@patch("connectors.sharepoint.requests.get")
def test_sharepoint_graph_dry_run(mock_get, mock_post, graph_env, tmp_path):
    mock_post.return_value = MagicMock(
        status_code=200,
        raise_for_status=lambda: None,
        json=lambda: {"access_token": "tok"},
    )
    mock_get.return_value = MagicMock(
        status_code=200,
        raise_for_status=lambda: None,
        json=lambda: {
            "value": [
                {"name": "policy.txt", "file": {}, "@microsoft.graph.downloadUrl": "http://x"},
            ]
        },
    )

    conn = SharePointGraphConnector()
    result = conn.sync(tmp_path / "out", dry_run=True)
    assert result.ok
    assert result.files_copied == 1
