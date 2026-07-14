"""Reciprocal Rank Fusion (RRF) for combining dense and sparse retrieval lists."""

from __future__ import annotations

from typing import Any, Sequence

from langchain_core.documents import Document

from rag.indexing import document_key


def reciprocal_rank_fusion(
    *ranked_lists: Sequence[Any],
    k: int,
    rrf_k: int = 60,
) -> list[Any]:
    """
    Fuse multiple ranked lists with RRF: score(d) = sum(1 / (rrf_k + rank)).

    Documents may be LangChain Document objects or any object with page_content
    and optional metadata.chunk_id.
    """
    if k <= 0:
        return []
    if not ranked_lists:
        return []

    scores: dict[str, float] = {}
    docs_by_key: dict[str, Any] = {}

    for ranked in ranked_lists:
        if not ranked:
            continue
        for rank, doc in enumerate(ranked, start=1):
            key = _key_for(doc)
            scores[key] = scores.get(key, 0.0) + 1.0 / (rrf_k + rank)
            docs_by_key[key] = doc

    if not scores:
        return []

    ordered = sorted(scores.items(), key=lambda item: item[1], reverse=True)
    return [docs_by_key[key] for key, _ in ordered[:k]]


def _key_for(doc: Any) -> str:
    if isinstance(doc, Document):
        return document_key(doc)
    meta = getattr(doc, "metadata", None) or {}
    chunk_id = meta.get("chunk_id")
    if chunk_id:
        return str(chunk_id)
    content = getattr(doc, "page_content", "") or ""
    return f"anon:{hash(content)}"
