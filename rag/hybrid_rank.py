"""Keyword rerank helper — re-exports from rag.rerank (legacy API).

For true hybrid retrieval use RAG_RETRIEVAL_MODE=hybrid (BM25 + dense + RRF).
"""

from typing import Any

from rag.rerank import keyword_overlap_score
from rag.rerank import rerank_documents as _rerank_documents


def rerank_documents(query: str, documents: list[Any], k: int) -> list[Any]:
    """Always keyword rerank (legacy hybrid_rank API)."""
    return _rerank_documents(query, documents, k, mode="keyword")


__all__ = ["keyword_overlap_score", "rerank_documents"]
