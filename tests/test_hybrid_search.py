"""Tests for hybrid retrieval mode (BM25 + dense + RRF)."""

import os
from unittest.mock import MagicMock, patch

from langchain_core.documents import Document
from rag.vector_store import reset_vector_store, retrieval_mode, search


def _doc(chunk_id: str, text: str) -> Document:
    return Document(page_content=text, metadata={"chunk_id": chunk_id})


@patch("rag.vector_store.get_vector_backend")
@patch("rag.vector_store.ensure_sparse_index")
def test_hybrid_search_fuses_dense_and_sparse(mock_sparse, mock_backend):
    os.environ["RAG_RETRIEVAL_MODE"] = "hybrid"
    try:
        dense_doc = _doc("dense-only", "Password reset link valid 24 hours.")
        sparse_doc = _doc("sparse-only", "VPN access request portal form.")
        shared = _doc("shared", "Shared policy about remote work.")

        backend = MagicMock()
        backend.similarity_search.return_value = [shared, dense_doc]
        mock_backend.return_value = backend

        sparse = MagicMock()
        sparse.search.return_value = [shared, sparse_doc]
        mock_sparse.return_value = sparse

        results = search("password reset VPN", domain_id="default", tenant_id="default", k=2)
        assert len(results) == 2
        ids = {d.metadata["chunk_id"] for d in results}
        assert "shared" in ids
    finally:
        os.environ.pop("RAG_RETRIEVAL_MODE", None)
        reset_vector_store()


def test_retrieval_mode_default_vector():
    os.environ.pop("RAG_RETRIEVAL_MODE", None)
    assert retrieval_mode() == "vector"


def test_retrieval_mode_hybrid():
    os.environ["RAG_RETRIEVAL_MODE"] = "hybrid"
    try:
        assert retrieval_mode() == "hybrid"
    finally:
        os.environ.pop("RAG_RETRIEVAL_MODE", None)
