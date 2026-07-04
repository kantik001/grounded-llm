"""Vector store backend interface (Chroma default, Qdrant optional)."""

from __future__ import annotations

from abc import ABC, abstractmethod
from typing import Any


class VectorBackend(ABC):
    """Pluggable vector index for RAG retrieval."""

    @abstractmethod
    def load(self, *, force_reindex: bool = False) -> None:
        """Open or rebuild the index."""

    @abstractmethod
    def similarity_search(
        self,
        query: str,
        *,
        k: int,
        domain_id: str,
        tenant_id: str,
    ) -> list[Any]:
        """Return LangChain Document-like objects with page_content and metadata."""

    @abstractmethod
    def index_stats_for_domain(self, domain_id: str, tenant_id: str) -> list[dict]:
        """Chunk counts per source file for admin index status."""

    @abstractmethod
    def reset(self) -> None:
        """Drop cached client handles (tests / hot reload)."""
