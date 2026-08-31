"""Tests for index-run collection naming helpers."""

from rag.kb.index_collections import chroma_run_dir, collection_name, run_suffix, sparse_run_dir


def test_run_suffix():
    run_id = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    assert run_suffix(run_id) == "a1b2c3d4e5f6"


def test_collection_name():
    name = collection_name("grounded_llm", "Acme", "HR", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
    assert name == "grounded_llm_acme_hr_a1b2c3d4e5f6"


def test_chroma_run_dir():
    path = chroma_run_dir("/data/chroma", "acme", "hr", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
    assert path.replace("\\", "/").endswith("runs/acme_hr_a1b2c3d4e5f6")
    assert path.startswith("/data/chroma") or "chroma" in path


def test_sparse_run_dir():
    path = sparse_run_dir("/data/sparse", "acme", "hr", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
    assert "runs" in path.replace("\\", "/")
    assert "acme_hr_a1b2c3d4e5f6" in path
