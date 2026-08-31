"""PostgreSQL pgvector backend (optional — set VECTOR_STORE=pgvector)."""

from __future__ import annotations

import os
import uuid
from typing import Any

from rag.embedding_cache import CachedHuggingFaceEmbeddings
from rag.indexing import split_file_documents, split_kb_documents
from rag.kb.index_collections import collection_name
from rag.vector_backend.base import VectorBackend
from rag.vector_backend.chroma_backend import EMBEDDING_MODEL, scan_kb_files


def normalize_pg_connection(url: str) -> str:
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
    return connection.replace("postgresql+psycopg://", "postgresql://", 1)


class PGVectorBackend(VectorBackend):
    """LangChain PGVector store. Requires: pip install -r api/requirements-pgvector.txt"""

    def __init__(self) -> None:
        self._scope_stores: dict[str, Any] = {}
        self._scope_collections: dict[str, str] = {}
        self._embeddings = CachedHuggingFaceEmbeddings(model_name=EMBEDDING_MODEL)
        self._collection_base = (
            os.environ.get("PGVECTOR_COLLECTION", "grounded_chunks").strip() or "grounded_chunks"
        )
        self._connection = pg_connection_url()

    def reset(self) -> None:
        self._scope_stores = {}
        self._scope_collections = {}

    def _scope_cache_key(self, tenant_id: str, domain_id: str, run_id: str) -> str:
        return f"{tenant_id}/{domain_id}/{run_id}"

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

        collection = collection_name(self._collection_base, tenant_id, domain_id, resolved)
        PGVector = self._pgvector_cls()
        store = PGVector(
            embeddings=self._embeddings,
            collection_name=collection,
            connection=self._connection,
            use_jsonb=True,
        )
        self._scope_stores[cache_key] = store
        self._scope_collections[cache_key] = collection
        return store

    def load(self, *, force_reindex: bool = False) -> None:
        if force_reindex or os.environ.get("FORCE_RAG_REINDEX", "").lower() in ("1", "true", "yes"):
            self.reset()
            self._index_all_scopes()

    def _index_all_scopes(self) -> None:
        documents = split_kb_documents()
        if not documents:
            return
        by_scope: dict[tuple[str, str], list] = {}
        for doc in documents:
            meta = doc.metadata or {}
            key = (str(meta.get("tenant_id") or "default"), str(meta.get("domain_id") or "default"))
            by_scope.setdefault(key, []).append(doc)
        PGVector = self._pgvector_cls()
        for (tenant, domain), scope_docs in by_scope.items():
            run_id = self.resolve_run_id(tenant, domain, for_write=True)
            collection = collection_name(self._collection_base, tenant, domain, run_id)
            store = self.open_scope(tenant, domain, run_id=run_id, for_write=True)
            try:
                store.delete_collection()
            except Exception:
                pass
            ids = [str(doc.metadata.get("chunk_id") or uuid.uuid4()) for doc in scope_docs]
            PGVector.from_documents(
                documents=scope_docs,
                embedding=self._embeddings,
                collection_name=collection,
                connection=self._connection,
                use_jsonb=True,
                ids=ids,
            )
            print(f"pgvector indexed {len(scope_docs)} chunks → {collection}")

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
        try:
            store.delete(filter={"tenant_id": tenant_id, "domain_id": domain_id, "filename": name})
        except Exception:
            pass
        chunks = split_file_documents(domain_id, path, tenant_id=tenant_id)
        if chunks:
            ids = [str(doc.metadata.get("chunk_id") or uuid.uuid4()) for doc in chunks]
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
        return store.similarity_search(
            query,
            k=k,
            filter={"domain_id": domain_id, "tenant_id": tenant_id},
        )

    def index_stats_for_domain(self, domain_id: str, tenant_id: str) -> list[dict]:
        run_id = self.resolve_run_id(tenant_id, domain_id)
        collection = self._scope_collections.get(
            self._scope_cache_key(tenant_id, domain_id, run_id),
            collection_name(self._collection_base, tenant_id, domain_id, run_id),
        )
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

        return [{"filename": filename or "unknown", "chunks": int(chunks)} for filename, chunks in rows]
