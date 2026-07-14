"""Retrieval reranking: keyword overlap (default hybrid) or cross-encoder (optional)."""

from __future__ import annotations

import os
import re
from typing import Any

_CROSS_ENCODER = None


def reranker_mode() -> str:
    """none | keyword | cross_encoder"""
    raw = (os.environ.get("RAG_RERANKER") or "").strip().lower()
    if raw in ("cross_encoder", "cross-encoder", "crossencoder"):
        return "cross_encoder"
    if raw in ("keyword", "hybrid"):
        return "keyword"
    if raw == "none":
        return "none"
    return "none"


def _tokens(text: str) -> set[str]:
    return set(re.findall(r"[a-z0-9]+", (text or "").lower()))


def keyword_overlap_score(query: str, text: str) -> float:
    q = _tokens(query)
    if not q:
        return 0.0
    return len(q & _tokens(text)) / len(q)


def _cross_encoder():
    global _CROSS_ENCODER
    if _CROSS_ENCODER is None:
        from sentence_transformers import CrossEncoder

        model_name = (
            os.environ.get("RAG_CROSS_ENCODER_MODEL") or "cross-encoder/ms-marco-MiniLM-L-6-v2"
        ).strip()
        _CROSS_ENCODER = CrossEncoder(model_name)
    return _CROSS_ENCODER


def cross_encoder_score(query: str, text: str) -> float:
    encoder = _cross_encoder()
    scores = encoder.predict([(query, text or "")])
    return float(scores[0])


def score_pair(query: str, text: str, mode: str) -> float:
    if mode == "cross_encoder":
        return cross_encoder_score(query, text)
    return keyword_overlap_score(query, text)


def rerank_documents(
    query: str,
    documents: list[Any],
    k: int,
    *,
    mode: str | None = None,
) -> list[Any]:
    """Rerank vector hits; mode from arg or RAG_RERANKER / RAG_RETRIEVAL_MODE."""
    mode = mode or reranker_mode()
    if mode == "none" or not documents or k <= 0:
        return documents[:k] if documents else []
    if len(documents) <= k:
        return documents[:k]

    scored: list[tuple[float, int, Any]] = []
    for rank, doc in enumerate(documents):
        text = getattr(doc, "page_content", "") or ""
        score = score_pair(query, text, mode)
        scored.append((score, -rank, doc))

    scored.sort(key=lambda item: (item[0], item[1]), reverse=True)
    return [doc for _, _, doc in scored[:k]]
