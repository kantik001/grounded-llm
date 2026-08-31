"""Qdrant vector backend (optional — set VECTOR_STORE=qdrant)."""

from __future__ import annotations

import os
import uuid
from typing import Any

from rag.embedding_cache import CachedHuggingFaceEmbeddings
from rag.indexing import split_file_documents, split_kb_documents
from rag.kb.index_collections import collection_name
from rag.vector_backend.base import VectorBackend
from rag.vector_backend.chroma_backend import EMBEDDING_MODEL, scan_kb_files


class QdrantBackend(VectorBackend):
    """LangChain Qdrant store. Requires: pip install -r api/requirements-qdrant.txt"""

    def __init__(self) -> None:
        self._scope_stores: dict[str, Any] = {}
        self._scope_collections: dict[str, str] = {}
        self._embeddings = CachedHuggingFaceEmbeddings(model_name=EMBEDDING_MODEL)
        self._collection_base = (
            os.environ.get("QDRANT_COLLECTION", "grounded_llm").strip() or "grounded_llm"
        )
        self._url = os.environ.get("QDRANT_URL", "http://127.0.0.1:6333").strip()

    def reset(self) -> None:
        self._scope_stores = {}
        self._scope_collections = {}

    def _scope_cache_key(self, tenant_id: str, domain_id: str, run_id: str) -> str:
        return f"{tenant_id}/{domain_id}/{run_id}"

    def _client_and_store(self, collection: str):
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
            collection_name=collection,
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

        collection = collection_name(self._collection_base, tenant_id, domain_id, resolved)
        client, store = self._client_and_store(collection)
        try:
            client.get_collection(collection)
        except Exception:
            pass
        self._scope_stores[cache_key] = store
        self._scope_collections[cache_key] = collection
        return store

    def load(self, *, force_reindex: bool = False) -> None:
        if force_reindex or os.environ.get("FORCE_RAG_REINDEX", "").lower() in ("1", "true", "yes"):
            self.reset()
            self._index_all_scopes()
            return

    def _index_all_scopes(self) -> None:
        documents = split_kb_documents()
        if not documents:
            print("No documents to index (Qdrant).")
            return
        by_scope: dict[tuple[str, str], list] = {}
        for doc in documents:
            meta = doc.metadata or {}
            key = (str(meta.get("tenant_id") or "default"), str(meta.get("domain_id") or "default"))
            by_scope.setdefault(key, []).append(doc)
        for (tenant, domain), scope_docs in by_scope.items():
            store = self.open_scope(tenant, domain, for_write=True)
            collection = self._scope_collections[self._scope_cache_key(tenant, domain, self.resolve_run_id(tenant, domain))]
            try:
                store.client.delete_collection(collection)
            except Exception:
                pass
            ids = [doc.metadata.get("chunk_id") or str(uuid.uuid4()) for doc in scope_docs]
            store.add_documents(scope_docs, ids=ids)
            print(f"Qdrant indexed {len(scope_docs)} chunks → {collection}")

    def refresh(self) -> dict:
        current = scan_kb_files()
        if not current:
            return {"mode": "full", "files": 0, "empty": True}
        self.load(force_reindex=True)
        return {"mode": "full", "files": len(current), "empty": False}

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
        name = filename or os.path.basename(path)
        resolved = self.resolve_run_id(tenant_id, domain_id, run_id=run_id, for_write=True)
        collection = self._scope_collections[self._scope_cache_key(tenant_id, domain_id, resolved)]
        try:
            store.client.delete(
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
        store = self.open_scope(tenant_id, domain_id, run_id=run_id)
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
        store = self.open_scope(tenant_id, domain_id, run_id=run_id)
        collection = self._scope_collections[self._scope_cache_key(tenant_id, domain_id, run_id)]
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
