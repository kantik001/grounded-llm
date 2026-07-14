"""Chroma vector backend (default reference implementation)."""

from __future__ import annotations

import os
import shutil
from typing import Any

from langchain_chroma import Chroma
from langchain_huggingface import HuggingFaceEmbeddings

from rag.indexing import split_kb_documents
from rag.vector_backend.base import VectorBackend

_PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
DEFAULT_PERSIST_DIR = os.path.join(_PROJECT_ROOT, "chroma_db")
EMBEDDING_MODEL = "intfloat/multilingual-e5-small"


def _persist_dir() -> str:
    return os.environ.get("CHROMA_PERSIST_DIR", DEFAULT_PERSIST_DIR).strip() or DEFAULT_PERSIST_DIR


class ChromaBackend(VectorBackend):
    def __init__(self) -> None:
        self._store: Chroma | None = None
        self._embeddings = HuggingFaceEmbeddings(model_name=EMBEDDING_MODEL)

    def reset(self) -> None:
        self._store = None

    def _create_store(self) -> Chroma | None:
        print("Creating vector store (Chroma)...")
        docs = split_kb_documents()
        if not docs:
            print("No documents to index.")
            return None
        print(f"Chunks: {len(docs)}")
        persist_dir = _persist_dir()
        store = Chroma.from_documents(docs, self._embeddings, persist_directory=persist_dir)
        print(f"Vector store saved to {persist_dir}")
        return store

    def load(self, *, force_reindex: bool = False) -> None:
        if self._store is not None and not force_reindex:
            return

        force = force_reindex or os.environ.get("FORCE_RAG_REINDEX", "").lower() in (
            "1",
            "true",
            "yes",
        )
        persist_dir = _persist_dir()

        if force and os.path.isdir(persist_dir):
            print("FORCE_RAG_REINDEX: removing old chroma_db")
            shutil.rmtree(persist_dir, ignore_errors=True)

        if os.path.exists(persist_dir) and os.listdir(persist_dir):
            self._store = Chroma(persist_directory=persist_dir, embedding_function=self._embeddings)
        else:
            self._store = self._create_store()

    def _filter(self, domain_id: str, tenant_id: str) -> dict:
        return {"$and": [{"domain_id": domain_id}, {"tenant_id": tenant_id}]}

    def similarity_search(
        self,
        query: str,
        *,
        k: int,
        domain_id: str,
        tenant_id: str,
    ) -> list[Any]:
        self.load()
        if self._store is None:
            return []
        return self._store.similarity_search(query, k=k, filter=self._filter(domain_id, tenant_id))

    def index_stats_for_domain(self, domain_id: str, tenant_id: str) -> list[dict]:
        self.load()
        if self._store is None:
            return []
        try:
            data = self._store._collection.get(  # noqa: SLF001
                where=self._filter(domain_id, tenant_id),
                include=["metadatas"],
            )
        except Exception:
            try:
                data = self._store._collection.get(
                    where={"domain_id": domain_id, "tenant_id": tenant_id},
                    include=["metadatas"],
                )
            except Exception:
                return []
        counts: dict[str, int] = {}
        for meta in data.get("metadatas") or []:
            if not meta:
                continue
            fn = meta.get("filename") or meta.get("source_file") or "unknown"
            counts[fn] = counts.get(fn, 0) + 1
        return [{"filename": name, "chunks": n} for name, n in sorted(counts.items())]
