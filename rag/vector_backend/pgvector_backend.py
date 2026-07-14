"""PostgreSQL pgvector backend (optional — set VECTOR_STORE=pgvector)."""

from __future__ import annotations

import os
import uuid
from typing import Any

from langchain_huggingface import HuggingFaceEmbeddings

from rag.indexing import split_kb_documents
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
        self._embeddings = HuggingFaceEmbeddings(model_name=EMBEDDING_MODEL)
        self._collection = (
            os.environ.get("PGVECTOR_COLLECTION", "grounded_chunks").strip() or "grounded_chunks"
        )
        self._connection = pg_connection_url()

    def reset(self) -> None:
        self._store = None

    def _pgvector_cls(self):
        try:
            from langchain_postgres import PGVector
        except ImportError as exc:
            raise RuntimeError(
                "pgvector backend requires optional deps: pip install -r api/requirements-pgvector.txt"
            ) from exc
        return PGVector

    def _open_store(self):
        PGVector = self._pgvector_cls()
        return PGVector(
            embeddings=self._embeddings,
            collection_name=self._collection,
            connection=self._connection,
            use_jsonb=True,
        )

    def _index_documents(self, documents: list[Any]) -> None:
        PGVector = self._pgvector_cls()
        if not documents:
            self._store = self._open_store()
            return
        ids = [str(doc.metadata.get("chunk_id") or uuid.uuid4()) for doc in documents]
        print(f"pgvector indexing chunks: {len(documents)}")
        self._store = PGVector.from_documents(
            documents=documents,
            embedding=self._embeddings,
            collection_name=self._collection,
            connection=self._connection,
            use_jsonb=True,
            ids=ids,
        )
        print(f"pgvector collection ready: {self._collection}")

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
        self.load()
        if self._store is None:
            return []
        return self._store.similarity_search(
            query,
            k=k,
            filter=self._metadata_filter(domain_id, tenant_id),
        )

    def index_stats_for_domain(self, domain_id: str, tenant_id: str) -> list[dict]:
        self.load()
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
                    cur.execute(sql, (self._collection, domain_id, tenant_id))
                    rows = cur.fetchall()
        except Exception:
            return []

        out: list[dict] = []
        for filename, chunks in rows:
            name = filename or "unknown"
            out.append({"filename": name, "chunks": int(chunks)})
        return out
