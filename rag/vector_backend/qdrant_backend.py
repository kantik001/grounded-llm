"""Qdrant vector backend (optional — set VECTOR_STORE=qdrant)."""

from __future__ import annotations

import os
import uuid
from typing import Any

from langchain_huggingface import HuggingFaceEmbeddings

from rag.indexing import split_kb_documents
from rag.vector_backend.base import VectorBackend
from rag.vector_backend.chroma_backend import EMBEDDING_MODEL

_PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))



class QdrantBackend(VectorBackend):
    """LangChain Qdrant store. Requires: pip install -r api/requirements-qdrant.txt"""

    def __init__(self) -> None:
        self._store = None
        self._embeddings = HuggingFaceEmbeddings(model_name=EMBEDDING_MODEL)
        self._collection = (
            os.environ.get("QDRANT_COLLECTION", "grounded_llm").strip() or "grounded_llm"
        )
        self._url = os.environ.get("QDRANT_URL", "http://127.0.0.1:6333").strip()

    def reset(self) -> None:
        self._store = None

    def _client_and_store(self):
        try:
            from langchain_qdrant import QdrantVectorStore
            from qdrant_client import QdrantClient
        except ImportError as exc:
            raise RuntimeError(
                "Qdrant backend requires optional deps: pip install -r api/requirements-qdrant.txt"
            ) from exc
        client = QdrantClient(url=self._url)
        return client, QdrantVectorStore(
            client=client,
            collection_name=self._collection,
            embeddings=self._embeddings,
        )

    def load(self, *, force_reindex: bool = False) -> None:
        if self._store is not None and not force_reindex:
            return

        force = force_reindex or os.environ.get("FORCE_RAG_REINDEX", "").lower() in (
            "1",
            "true",
            "yes",
        )
        client, store = self._client_and_store()

        if force:
            try:
                client.delete_collection(self._collection)
            except Exception:
                pass

        try:
            client.get_collection(self._collection)
            self._store = store
            return
        except Exception:
            pass

        documents = split_kb_documents()
        if not documents:
            print("No documents to index (Qdrant).")
            self._store = store
            return

        print(f"Qdrant indexing chunks: {len(documents)}")
        ids = [doc.metadata.get("chunk_id") or str(uuid.uuid4()) for doc in documents]
        store.add_documents(documents, ids=ids)
        self._store = store
        print(f"Qdrant collection ready: {self._collection} @ {self._url}")

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
        flt = {"must": [{"key": "domain_id", "match": {"value": domain_id}},
                        {"key": "tenant_id", "match": {"value": tenant_id}}]}
        try:
            return self._store.similarity_search(query, k=k, filter=flt)
        except TypeError:
            return self._store.similarity_search(
                query,
                k=k,
                filter={"domain_id": domain_id, "tenant_id": tenant_id},
            )

    def index_stats_for_domain(self, domain_id: str, tenant_id: str) -> list[dict]:
        self.load()
        if self._store is None:
            return []
        try:
            client = self._store.client
            scroll_filter = {
                "must": [
                    {"key": "domain_id", "match": {"value": domain_id}},
                    {"key": "tenant_id", "match": {"value": tenant_id}},
                ]
            }
            counts: dict[str, int] = {}
            offset = None
            while True:
                points, offset = client.scroll(
                    collection_name=self._collection,
                    scroll_filter=scroll_filter,
                    limit=256,
                    offset=offset,
                    with_payload=True,
                )
                if not points:
                    break
                for point in points:
                    payload = point.payload or {}
                    meta = payload.get("metadata") or payload
                    fn = meta.get("filename") or meta.get("source_file") or "unknown"
                    counts[fn] = counts.get(fn, 0) + 1
                if offset is None:
                    break
        except Exception:
            return []
        return [{"filename": name, "chunks": n} for name, n in sorted(counts.items())]
