"""Tests for reciprocal rank fusion."""

from langchain_core.documents import Document
from rag.rrf import reciprocal_rank_fusion


def _doc(chunk_id: str, text: str = "") -> Document:
    return Document(page_content=text or chunk_id, metadata={"chunk_id": chunk_id})


def test_rrf_prefers_items_in_both_lists():
    dense = [_doc("a", "alpha"), _doc("b", "beta"), _doc("c", "gamma")]
    sparse = [_doc("b", "beta"), _doc("d", "delta"), _doc("a", "alpha")]
    fused = reciprocal_rank_fusion(dense, sparse, k=2)
    ids = [d.metadata["chunk_id"] for d in fused]
    assert ids[0] in ("a", "b")
    assert len(ids) == 2


def test_rrf_respects_k():
    dense = [_doc("x"), _doc("y"), _doc("z")]
    sparse = [_doc("y"), _doc("z"), _doc("w")]
    fused = reciprocal_rank_fusion(dense, sparse, k=1)
    assert len(fused) == 1


def test_rrf_single_list():
    dense = [_doc("only")]
    fused = reciprocal_rank_fusion(dense, k=3)
    assert len(fused) == 1
    assert fused[0].metadata["chunk_id"] == "only"


def test_rrf_empty_returns_empty():
    assert reciprocal_rank_fusion([], k=5) == []
    assert reciprocal_rank_fusion([], [], k=5) == []
