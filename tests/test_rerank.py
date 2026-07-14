"""Tests for retrieval reranking."""

import os
from types import SimpleNamespace

from rag.rerank import keyword_overlap_score, rerank_documents, reranker_mode


def _doc(text: str):
    return SimpleNamespace(page_content=text, metadata={})


def test_reranker_mode_default_none():
    os.environ.pop("RAG_RERANKER", None)
    os.environ.pop("RAG_RETRIEVAL_MODE", None)
    assert reranker_mode() == "none"


def test_reranker_mode_hybrid_no_implicit_keyword():
    os.environ.pop("RAG_RERANKER", None)
    os.environ["RAG_RETRIEVAL_MODE"] = "hybrid"
    try:
        assert reranker_mode() == "none"
    finally:
        os.environ.pop("RAG_RETRIEVAL_MODE", None)


def test_reranker_mode_explicit_cross_encoder():
    os.environ["RAG_RERANKER"] = "cross_encoder"
    try:
        assert reranker_mode() == "cross_encoder"
    finally:
        os.environ.pop("RAG_RERANKER", None)


def test_keyword_rerank_promotes_match():
    docs = [
        _doc("IT portal hours 08:00 to 18:00."),
        _doc("Password reset link valid for 24 hours."),
    ]
    ranked = rerank_documents("password reset link", docs, k=1, mode="keyword")
    assert "24 hours" in ranked[0].page_content


def test_keyword_overlap_full_match():
    assert keyword_overlap_score("vpn access", "Request VPN access via portal.") == 1.0
