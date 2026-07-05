"""Tests for ingest connectors."""

from connectors.local_folder import LocalFolderConnector


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
