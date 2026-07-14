"""Vector store backend factory."""

from __future__ import annotations

import os

from rag.vector_backend.base import VectorBackend
from rag.vector_backend.chroma_backend import ChromaBackend

_backend: VectorBackend | None = None


def get_vector_backend() -> VectorBackend:
    global _backend
    if _backend is not None:
        return _backend

    name = (os.environ.get("VECTOR_STORE") or "chroma").strip().lower()
    if name in ("chroma", ""):
        _backend = ChromaBackend()
    elif name == "qdrant":
        from rag.vector_backend.qdrant_backend import QdrantBackend

        _backend = QdrantBackend()
    elif name == "pgvector":
        from rag.vector_backend.pgvector_backend import PGVectorBackend

        _backend = PGVectorBackend()
    else:
        raise ValueError(
            f"Unknown VECTOR_STORE={name!r} (supported: chroma, qdrant, pgvector)"
        )
    return _backend


def reset_vector_backend() -> None:
    global _backend
    if _backend is not None:
        _backend.reset()
    _backend = None
