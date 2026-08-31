"""Tests for BM25 sparse index."""

import os
import tempfile

from langchain_core.documents import Document
from rag.sparse_index import BM25SparseIndex, ensure_sparse_index, reset_sparse_index

_RUN_ID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"


def _chunk(tenant: str, domain: str, filename: str, seq: int, text: str) -> Document:
    return Document(
        page_content=text,
        metadata={
            "tenant_id": tenant,
            "domain_id": domain,
            "filename": filename,
            "chunk_id": f"{tenant}/{domain}/{filename}/{seq}",
        },
    )


def test_bm25_search_finds_keyword_match():
    reset_sparse_index()
    idx = BM25SparseIndex(tenant_id="default", domain_id="default", run_id=_RUN_ID)
    idx.build(
        [
            _chunk("default", "default", "a.txt", 0, "IT portal hours are 08:00 to 18:00."),
            _chunk("default", "default", "b.txt", 0, "Password reset link valid for 24 hours."),
        ],
        persist=False,
    )
    hits = idx.search(
        "password reset link",
        domain_id="default",
        tenant_id="default",
        k=1,
    )
    assert len(hits) == 1
    assert "24 hours" in hits[0].page_content


def test_bm25_scoped_by_domain():
    reset_sparse_index()
    idx_it = BM25SparseIndex(tenant_id="default", domain_id="it", run_id=_RUN_ID)
    idx_it.build(
        [_chunk("default", "it", "b.txt", 0, "VPN access request via portal.")],
        persist=False,
    )
    hits = idx_it.search("VPN access", domain_id="it", tenant_id="default", k=1)
    assert len(hits) == 1
    assert "VPN" in hits[0].page_content

    idx_hr = BM25SparseIndex(tenant_id="default", domain_id="hr", run_id=_RUN_ID)
    idx_hr.build(
        [_chunk("default", "hr", "a.txt", 0, "Vacation policy allows 28 days.")],
        persist=False,
    )
    assert idx_hr.search("VPN access", domain_id="hr", tenant_id="default", k=1) == []


def test_ensure_sparse_index_force_rebuild(monkeypatch, tmp_path):
    reset_sparse_index()
    monkeypatch.setenv("SPARSE_INDEX_DIR", str(tmp_path))
    monkeypatch.setattr("rag.sparse_index.resolve_run_id", lambda _t, _d: _RUN_ID)
    monkeypatch.setattr(
        "rag.sparse_index.split_kb_documents",
        lambda: [_chunk("default", "default", "a.txt", 0, "Annual leave is 28 days.")],
    )
    idx = ensure_sparse_index(tenant_id="default", domain_id="default", force_reindex=True)
    assert idx.is_ready()
    hits = idx.search("annual leave", domain_id="default", tenant_id="default", k=1)
    assert len(hits) == 1


def test_bm25_persist_and_load():
    reset_sparse_index()
    with tempfile.TemporaryDirectory() as tmp:
        os.environ["SPARSE_INDEX_DIR"] = tmp
        try:
            idx = BM25SparseIndex(tenant_id="default", domain_id="default", run_id=_RUN_ID)
            idx.build(
                [_chunk("default", "default", "a.txt", 0, "Annual leave is 28 days.")],
                persist=True,
            )
            loaded = BM25SparseIndex(tenant_id="default", domain_id="default", run_id=_RUN_ID)
            assert loaded.load() is True
            hits = loaded.search("annual leave", domain_id="default", tenant_id="default", k=1)
            assert len(hits) == 1
            assert "28" in hits[0].page_content
        finally:
            os.environ.pop("SPARSE_INDEX_DIR", None)
            reset_sparse_index()
