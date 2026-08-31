"""PostgreSQL pgvector backend (optional — set VECTOR_STORE=pgvector)."""

from __future__ import annotations

import os
import uuid
from typing import Any

from rag.embedding_cache import CachedHuggingFaceEmbeddings
from rag.indexing import split_file_documents, split_kb_documents
from rag.kb.index_collections import collection_name
from rag.vector_backend.base import VectorBackend
from rag.vector_backend.chroma_backend import EMBEDDING_MODEL


def normalize_pg_connection(url: str) -> str:
    """Convert postgres:// or postgresql:// to postgresql+psycopg:// for langchain-postgres."""
    raw = (url or "").strip()
    if not raw:
        return ""
    if raw.startswith("postgresql+psycopg://"):
        return raw
    if raw.startswith("postgres://"):
        return "postgresql+psycopg://" + raw[len("postgres://") :]
    if raw.startswith("postgresql://"):
        return "postgresql+psycopg://" + raw[len("postgresql://") :]
    return raw


def pg_connection_url() -> str:
    url = os.environ.get("PGVECTOR_URL") or os.environ.get("DATABASE_URL") or ""
    conn = normalize_pg_connection(url)
    if not conn:
        raise RuntimeError(
            "VECTOR_STORE=pgvector requires PGVECTOR_URL or DATABASE_URL "
            "(postgresql+psycopg://...)"
        )
    return conn


def psycopg_dsn(connection: str) -> str:
    """DSN for psycopg (without SQLAlchemy driver suffix)."""
    return connection.replace("postgresql+psycopg://", "postgresql://", 1)


class PGVectorBackend(VectorBackend):
    """LangChain PGVector store. Requires: pip install -r api/requirements-pgvector.txt"""

    def __init__(self) -> None:
        self._store = None
        self._scope_stores: dict[str, Any] = {}
        self._scope_collections: dict[str, str] = {}
        self._embeddings = CachedHuggingFaceEmbeddings(model_name=EMBEDDING_MODEL)
        self._collection_base = (
            os.environ.get("PGVECTOR_COLLECTION", "grounded_chunks").strip() or "grounded_chunks"
        )
        self._connection = pg_connection_url()

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

    def _pgvector_cls(self):
        try:
            from langchain_postgres import PGVector
        except ImportError as exc:
            raise RuntimeError(
                "pgvector backend requires optional deps: pip install -r api/requirements-pgvector.txt"
            ) from exc
        return PGVector

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
        PGVector = self._pgvector_cls()
        store = PGVector(
            embeddings=self._embeddings,
            collection_name=collection,
            connection=self._connection,
            use_jsonb=True,
        )
        self._scope_stores[cache_key] = store
        self._scope_collections[cache_key] = collection
        if resolved is None:
            self._store = store
        return store

    def _open_store(self, collection: str | None = None):
        PGVector = self._pgvector_cls()
        name = collection or self._collection_base
        return PGVector(
            embeddings=self._embeddings,
            collection_name=name,
            connection=self._connection,
            use_jsonb=True,
        )

    def _index_documents(self, documents: list[Any], collection: str | None = None) -> None:
        PGVector = self._pgvector_cls()
        name = collection or self._collection_base
        if not documents:
            self._store = self._open_store(name)
            return
        ids = [str(doc.metadata.get("chunk_id") or uuid.uuid4()) for doc in documents]
        print(f"pgvector indexing chunks: {len(documents)}")
        self._store = PGVector.from_documents(
            documents=documents,
            embedding=self._embeddings,
            collection_name=name,
            connection=self._connection,
            use_jsonb=True,
            ids=ids,
        )
        print(f"pgvector collection ready: {name}")

    def load(self, *, force_reindex: bool = False) -> None:
        if self._store is not None and not force_reindex:
            return

        force = force_reindex or os.environ.get("FORCE_RAG_REINDEX", "").lower() in (
            "1",
            "true",
            "yes",
        )

        if force:
            store = self._open_store()
            try:
                store.delete_collection()
            except Exception:
                pass
            self._index_documents(split_kb_documents())
            return

        self._store = self._open_store()

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
        try:
            store.delete(filter={"tenant_id": tenant_id, "domain_id": domain_id, "filename": name})
        except Exception:
            pass
        chunks = split_file_documents(domain_id, path, tenant_id=tenant_id)
        if chunks:
            ids = [str(doc.metadata.get("chunk_id") or uuid.uuid4()) for doc in chunks]
            store.add_documents(chunks, ids=ids)
        return len(chunks)

    def _metadata_filter(self, domain_id: str, tenant_id: str) -> dict[str, str]:
        return {"domain_id": domain_id, "tenant_id": tenant_id}

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
        return store.similarity_search(
            query,
            k=k,
            filter=self._metadata_filter(domain_id, tenant_id),
        )

    def index_stats_for_domain(self, domain_id: str, tenant_id: str) -> list[dict]:
        run_id = self.resolve_run_id(tenant_id, domain_id)
        cache_key = self._scope_cache_key(tenant_id, domain_id, run_id)
        collection = self._scope_collections.get(cache_key, self._collection_base)
        try:
            import psycopg
        except ImportError:
            return []

        sql = """
            SELECT e.cmetadata->>'filename' AS filename, COUNT(*)::int AS chunks
            FROM langchain_pg_embedding e
            JOIN langchain_pg_collection c ON e.collection_id = c.uuid
            WHERE c.name = %s
              AND e.cmetadata->>'domain_id' = %s
              AND e.cmetadata->>'tenant_id' = %s
            GROUP BY 1
            ORDER BY 1
        """
        try:
            with psycopg.connect(psycopg_dsn(self._connection)) as conn:
                with conn.cursor() as cur:
                    cur.execute(sql, (collection, domain_id, tenant_id))
                    rows = cur.fetchall()
        except Exception:
            return []

        out: list[dict] = []
        for filename, chunks in rows:
            name = filename or "unknown"
            out.append({"filename": name, "chunks": int(chunks)})
        return out
