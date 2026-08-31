"""Qdrant vector backend (optional — set VECTOR_STORE=qdrant)."""

from __future__ import annotations

import os
import uuid
from typing import Any

from rag.embedding_cache import CachedHuggingFaceEmbeddings
from rag.indexing import split_file_documents, split_kb_documents
from rag.kb.index_collections import collection_name
from rag.vector_backend.base import VectorBackend
from rag.vector_backend.chroma_backend import EMBEDDING_MODEL


class QdrantBackend(VectorBackend):
    """LangChain Qdrant store. Requires: pip install -r api/requirements-qdrant.txt"""

    def __init__(self) -> None:
        self._store = None
        self._scope_stores: dict[str, Any] = {}
        self._scope_collections: dict[str, str] = {}
        self._embeddings = CachedHuggingFaceEmbeddings(model_name=EMBEDDING_MODEL)
        self._collection_base = (
            os.environ.get("QDRANT_COLLECTION", "grounded_llm").strip() or "grounded_llm"
        )
        self._url = os.environ.get("QDRANT_URL", "http://127.0.0.1:6333").strip()

    def reset(self) -> None:
        self._store = None
        self._scope_stores = {}
        self._scope_collections = {}

    def _scope_cache_key(self, tenant_id: str, domain_id: str, run_id: str | None) -> str:
        return f"{tenant_id}/{domain_id}/{run_id or 'legacy'}"

    def _resolved_collection(self, tenant_id: str, domain_id: str, run_id: str | None) -> str:
        if run_id:
            return collection_name(self._collection_base, tenant_id, domain_id, run_id)
        return self._collection_base

    def _client_and_store(self, collection: str | None = None):
        try:
            from langchain_qdrant import QdrantVectorStore
            from qdrant_client import QdrantClient
        except ImportError as exc:
            raise RuntimeError(
                "Qdrant backend requires optional deps: pip install -r api/requirements-qdrant.txt"
            ) from exc
        name = collection or self._collection_base
        client = QdrantClient(url=self._url)
        return client, QdrantVectorStore(
            client=client,
            collection_name=name,
            embeddings=self._embeddings,
        )

    def open_scope(
        self,
        tenant_id: str,
        domain_id: str,
        *,
        run_id: str | None = None,
        for_write: bool = False,
    ):
        resolved = self.resolve_run_id(tenant_id, domain_id, run_id=run_id, for_write=for_write)
        cache_key = self._scope_cache_key(tenant_id, domain_id, resolved)
        if cache_key in self._scope_stores:
            return self._scope_stores[cache_key]

        collection = self._resolved_collection(tenant_id, domain_id, resolved)
        client, store = self._client_and_store(collection)
        try:
            client.get_collection(collection)
        except Exception:
            pass
        self._scope_stores[cache_key] = store
        self._scope_collections[cache_key] = collection
        if resolved is None:
            self._store = store
        return store

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
                client.delete_collection(self._collection_base)
            except Exception:
                pass

        try:
            client.get_collection(self._collection_base)
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
        print(f"Qdrant collection ready: {self._collection_base} @ {self._url}")

    def upsert_kb_file(
        self,
        tenant_id: str,
        domain_id: str,
        path: str,
        *,
        filename: str | None = None,
        run_id: str | None = None,
    ) -> int:
        store = self.open_scope(tenant_id, domain_id, run_id=run_id, for_write=True)
        if store is None:
            return 0
        name = filename or os.path.basename(path)
        cache_key = self._scope_cache_key(
            tenant_id,
            domain_id,
            self.resolve_run_id(tenant_id, domain_id, run_id=run_id, for_write=True),
        )
        collection = self._scope_collections.get(cache_key, self._collection_base)
        client = store.client
        try:
            client.delete(
                collection_name=collection,
                points_selector={
                    "filter": {
                        "must": [
                            {"key": "tenant_id", "match": {"value": tenant_id}},
                            {"key": "domain_id", "match": {"value": domain_id}},
                            {"key": "filename", "match": {"value": name}},
                        ]
                    }
                },
            )
        except Exception:
            pass
        chunks = split_file_documents(domain_id, path, tenant_id=tenant_id)
        if chunks:
            ids = [doc.metadata.get("chunk_id") or str(uuid.uuid4()) for doc in chunks]
            store.add_documents(chunks, ids=ids)
        return len(chunks)

    def similarity_search(
        self,
        query: str,
        *,
        k: int,
        domain_id: str,
        tenant_id: str,
    ) -> list[Any]:
        run_id = self.resolve_run_id(tenant_id, domain_id)
        if run_id:
            store = self.open_scope(tenant_id, domain_id, run_id=run_id)
        else:
            self.load()
            store = self._store
        if store is None:
            return []
        flt = {
            "must": [
                {"key": "domain_id", "match": {"value": domain_id}},
                {"key": "tenant_id", "match": {"value": tenant_id}},
            ]
        }
        try:
            return store.similarity_search(query, k=k, filter=flt)
        except TypeError:
            return store.similarity_search(
                query,
                k=k,
                filter={"domain_id": domain_id, "tenant_id": tenant_id},
            )

    def index_stats_for_domain(self, domain_id: str, tenant_id: str) -> list[dict]:
        run_id = self.resolve_run_id(tenant_id, domain_id)
        if run_id:
            store = self.open_scope(tenant_id, domain_id, run_id=run_id)
        else:
            self.load()
            store = self._store
        if store is None:
            return []
        cache_key = self._scope_cache_key(tenant_id, domain_id, run_id)
        collection = self._scope_collections.get(cache_key, self._collection_base)
        try:
            client = store.client
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
                    collection_name=collection,
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
