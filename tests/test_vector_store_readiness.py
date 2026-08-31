"""Tests for vector store readiness probe."""

import pytest


@pytest.fixture(autouse=True)
def _chroma_backend(monkeypatch, tmp_path):
    monkeypatch.setenv("VECTOR_STORE", "chroma")
    monkeypatch.setenv("CHROMA_PERSIST_DIR", str(tmp_path))


def test_readiness_pending_without_runs_dir():
    from rag.vector_store import readiness_index_check

    label, ok = readiness_index_check()
    assert label == "pending"
    assert ok is True


def test_readiness_empty_runs_dir(tmp_path, monkeypatch):
    runs = tmp_path / "runs"
    runs.mkdir()
    monkeypatch.setenv("CHROMA_PERSIST_DIR", str(tmp_path))

    from rag.vector_store import readiness_index_check

    label, ok = readiness_index_check()
    assert label == "empty"
    assert ok is True


def test_readiness_ok_when_run_data_exists(tmp_path, monkeypatch):
    run_dir = tmp_path / "runs" / "default_default_abc123"
    run_dir.mkdir(parents=True)
    (run_dir / "chunks").mkdir()
    monkeypatch.setenv("CHROMA_PERSIST_DIR", str(tmp_path))

    from rag.vector_store import readiness_index_check

    label, ok = readiness_index_check()
    assert label == "ok"
    assert ok is True
