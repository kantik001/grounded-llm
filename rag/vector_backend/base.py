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

    def upsert_kb_file(
        self,
        tenant_id: str,
        domain_id: str,
        path: str,
        *,
        filename: str | None = None,
        run_id: str | None = None,
    ) -> int:
        """Re-embed one KB file into the scoped index run."""
        raise NotImplementedError

    def refresh(self) -> dict:
        """Sync the index with files on disk.

        Backends may override with an incremental implementation (see
        ChromaBackend); the default is a full rebuild.
        """
        self.load(force_reindex=True)
        return {"mode": "full"}

    def resolve_run_id(
        self,
        tenant_id: str,
        domain_id: str,
        *,
        run_id: str | None = None,
        for_write: bool = False,
    ) -> str:
        from rag.kb.index_runs import resolve_run_id as kb_resolve_run_id

        if run_id:
            return run_id
        if for_write:
            return kb_resolve_run_id(tenant_id, domain_id)
        return kb_resolve_run_id(tenant_id, domain_id)
