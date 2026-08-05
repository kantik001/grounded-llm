"""Tests for ingest connectors."""

import pytest
from connectors.local_folder import LocalFolderConnector
from connectors.registry import CONNECTOR_NAMES, build_connector, requires_source


def test_registry_lists_expected_connectors():
    assert "local_folder" in CONNECTOR_NAMES
    assert "google_drive" in CONNECTOR_NAMES
    assert "confluence" in CONNECTOR_NAMES
    assert requires_source("local_folder") is True
    assert requires_source("confluence") is False


def test_build_connector_local_folder(tmp_path):
    src = tmp_path / "src"
    src.mkdir()
    conn = build_connector("local_folder", str(src))
    assert conn.name == "local_folder"


def test_build_connector_unknown():
    with pytest.raises(ValueError, match="Unknown connector"):
        build_connector("not_a_real_connector")


def test_local_folder_dry_run(tmp_path):
    src = tmp_path / "src"
    src.mkdir()
    (src / "policy.txt").write_text("Vacation days: 28", encoding="utf-8")
    (src / "skip.bin").write_bytes(b"\x00")

    dest = tmp_path / "dest"
    conn = LocalFolderConnector(src)
    result = conn.sync(dest, dry_run=True)

    assert result.ok
    assert result.files_copied == 1
    assert result.files_skipped == 1
    assert not dest.exists() or not list(dest.iterdir())


def test_local_folder_copies_files(tmp_path):
    src = tmp_path / "src"
    src.mkdir()
    (src / "a.txt").write_text("hello", encoding="utf-8")

    dest = tmp_path / "dest"
    conn = LocalFolderConnector(src)
    result = conn.sync(dest)

    assert result.ok
    assert (dest / "a.txt").read_text(encoding="utf-8") == "hello"
