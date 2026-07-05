"""Tests for Google Drive connector (mocked API)."""

from unittest.mock import MagicMock, patch

import pytest
from connectors.google_drive import GoogleDriveConnector


@pytest.fixture
def drive_env(tmp_path, monkeypatch):
    creds = tmp_path / "sa.json"
    creds.write_text('{"type":"service_account"}', encoding="utf-8")
    monkeypatch.setenv("GOOGLE_APPLICATION_CREDENTIALS", str(creds))
    monkeypatch.setenv("GOOGLE_DRIVE_FOLDER_ID", "folder123")


def test_google_drive_missing_folder(tmp_path, monkeypatch):
    creds = tmp_path / "sa.json"
    creds.write_text("{}", encoding="utf-8")
    monkeypatch.setenv("GOOGLE_APPLICATION_CREDENTIALS", str(creds))
    monkeypatch.delenv("GOOGLE_DRIVE_FOLDER_ID", raising=False)
    with pytest.raises(ValueError, match="GOOGLE_DRIVE_FOLDER_ID"):
        GoogleDriveConnector()


@patch("connectors.google_drive.GoogleDriveConnector._drive_service")
def test_google_drive_dry_run(mock_service, drive_env, tmp_path):
    service = MagicMock()
    mock_service.return_value = service
    service.files.return_value.list.return_value.execute.return_value = {
        "files": [
            {"id": "1", "name": "policy.txt", "mimeType": "text/plain"},
        ]
    }

    conn = GoogleDriveConnector()
    result = conn.sync(tmp_path / "out", dry_run=True)
    assert result.ok
    assert result.files_copied == 1
