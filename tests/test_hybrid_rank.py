"""Tests for keyword hybrid reranking."""

from types import SimpleNamespace

from rag.hybrid_rank import keyword_overlap_score, rerank_documents


def _doc(text: str):
    return SimpleNamespace(page_content=text, metadata={})


def test_keyword_overlap_full_match():
    assert keyword_overlap_score("password reset link", "The password reset link expires in 24 hours.") == 1.0


def test_keyword_overlap_partial():
    score = keyword_overlap_score("vpn access request", "Submit the VPN Access form in the IT Portal.")
    assert 0.0 < score < 1.0


def test_rerank_promotes_keyword_match():
    docs = [
        _doc("General IT portal hours are 08:00 to 18:00."),
        _doc("The password reset link expires in 24 hours."),
    ]
    ranked = rerank_documents("password reset link valid", docs, k=1)
    assert "24 hours" in ranked[0].page_content


def test_rerank_returns_k():
    docs = [_doc(f"doc {i} with policy text") for i in range(5)]
    ranked = rerank_documents("policy text", docs, k=3)
    assert len(ranked) == 3
