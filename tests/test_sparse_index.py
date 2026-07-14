"""Tests for BM25 sparse index."""

import os
import tempfile

from langchain_core.documents import Document
from rag.sparse_index import BM25SparseIndex, reset_sparse_index


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
    idx = BM25SparseIndex()
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
    idx = BM25SparseIndex()
    idx.build(
        [
            _chunk("default", "hr", "a.txt", 0, "Vacation policy allows 28 days."),
            _chunk("default", "it", "b.txt", 0, "VPN access request via portal."),
        ],
        persist=False,
    )
    hits = idx.search("VPN access", domain_id="it", tenant_id="default", k=1)
    assert len(hits) == 1
    assert "VPN" in hits[0].page_content


def test_bm25_persist_and_load():
    reset_sparse_index()
    with tempfile.TemporaryDirectory() as tmp:
        os.environ["SPARSE_INDEX_DIR"] = tmp
        try:
            idx = BM25SparseIndex()
            idx.build(
                [_chunk("default", "default", "a.txt", 0, "Annual leave is 28 days.")],
                persist=True,
            )
            loaded = BM25SparseIndex()
            assert loaded.load() is True
            hits = loaded.search("annual leave", domain_id="default", tenant_id="default", k=1)
            assert len(hits) == 1
            assert "28" in hits[0].page_content
        finally:
            os.environ.pop("SPARSE_INDEX_DIR", None)
            reset_sparse_index()
