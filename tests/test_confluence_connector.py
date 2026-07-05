"""Tests for Confluence REST connector (mocked HTTP)."""

from unittest.mock import MagicMock, patch

import pytest
from connectors.confluence import ConfluenceConnector


@pytest.fixture
def confluence_env(monkeypatch):
    monkeypatch.setenv("CONFLUENCE_BASE_URL", "https://example.atlassian.net/wiki")
    monkeypatch.setenv("CONFLUENCE_EMAIL", "bot@example.com")
    monkeypatch.setenv("CONFLUENCE_API_TOKEN", "token")
    monkeypatch.setenv("CONFLUENCE_SPACE_KEY", "HR")


def test_confluence_missing_config():
    with pytest.raises(ValueError, match="Missing Confluence config"):
        ConfluenceConnector(base_url="", email="", api_token="", space_key="")


@patch("connectors.confluence.requests.Session")
def test_confluence_dry_run(mock_session_cls, confluence_env, tmp_path):
    sess = MagicMock()
    mock_session_cls.return_value = sess

    page_resp = MagicMock()
    page_resp.json.return_value = {
        "results": [
            {
                "id": "1",
                "title": "Vacation Policy",
                "body": {"storage": {"value": "<p>28 days vacation</p>"}},
            }
        ],
        "_links": {},
    }
    page_resp.raise_for_status = lambda: None

    att_resp = MagicMock()
    att_resp.json.return_value = {"results": []}
    att_resp.raise_for_status = lambda: None

    sess.get.side_effect = [page_resp, att_resp]

    conn = ConfluenceConnector()
    result = conn.sync(tmp_path / "out", dry_run=True)
    assert result.ok
    assert result.files_copied == 1
