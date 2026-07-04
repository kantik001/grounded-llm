"""Keyword hybrid reranking for retrieval (no extra ML deps)."""

from __future__ import annotations

import re
from typing import Any


def _tokens(text: str) -> set[str]:
    return set(re.findall(r"[a-z0-9]+", (text or "").lower()))


def keyword_overlap_score(query: str, text: str) -> float:
    q = _tokens(query)
    if not q:
        return 0.0
    overlap = len(q & _tokens(text))
    return overlap / len(q)


def rerank_documents(query: str, documents: list[Any], k: int) -> list[Any]:
    """Rerank vector hits by blending vector order with keyword overlap."""
    if not documents or k <= 0:
        return []
    if len(documents) <= k:
        return documents[:k]

    scored: list[tuple[float, int, Any]] = []
    for rank, doc in enumerate(documents):
        text = getattr(doc, "page_content", "") or ""
        score = keyword_overlap_score(query, text)
        scored.append((score, -rank, doc))

    scored.sort(key=lambda item: (item[0], item[1]), reverse=True)
    return [doc for _, _, doc in scored[:k]]
