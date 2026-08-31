"""Disposable index runs (blue/green index generations)."""

from __future__ import annotations

import json
import os
import uuid
from dataclasses import dataclass
from typing import Any

from rag.kb.documents import _connect
from rag.vector_backend.chroma_backend import EMBEDDING_MODEL, embedding_signature


@dataclass
class IndexRun:
    id: str
    tenant_id: str
    domain_id: str
    backend: str
    embedding_model: str
    chunk_schema: dict[str, Any]
    status: str


def default_chunk_schema() -> dict[str, Any]:
    return {"chunk_size": 500, "chunk_overlap": 50, "splitter": "recursive", "schema": 1}


def index_profile() -> dict[str, Any]:
    sig = embedding_signature()
    return {
        "backend": (os.environ.get("VECTOR_STORE") or "chroma").strip().lower(),
        "embedding_model": sig.get("model") or EMBEDDING_MODEL,
        "chunk_schema": default_chunk_schema(),
        "e5_prefixes": sig.get("e5_prefixes"),
        "schema": sig.get("schema", 1),
    }


def create_index_run(tenant_id: str, domain_id: str) -> str:
    profile = index_profile()
    run_id = str(uuid.uuid4())
    with _connect() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO index_runs (id, tenant_id, domain_id, backend, embedding_model, chunk_schema, status)
                VALUES (%s, NULLIF(%s, ''), NULLIF(%s, ''), %s, %s, %s::jsonb, 'building')
                """,
                (
                    run_id,
                    tenant_id,
                    domain_id,
                    profile["backend"],
                    profile["embedding_model"],
                    json.dumps(profile["chunk_schema"]),
                ),
            )
        conn.commit()
    return run_id


def activate_index_run(tenant_id: str, domain_id: str, run_id: str) -> None:
    with _connect() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                UPDATE index_runs SET status = 'retired'
                WHERE tenant_id IS NOT DISTINCT FROM NULLIF(%s, '')::text
                  AND domain_id IS NOT DISTINCT FROM NULLIF(%s, '')::text
                  AND status = 'active' AND id <> %s
                """,
                (tenant_id, domain_id, run_id),
            )
            cur.execute(
                "UPDATE index_runs SET status = 'active', activated_at = NOW() WHERE id = %s",
                (run_id,),
            )
            cur.execute(
                """
                INSERT INTO index_run_active (tenant_id, domain_id, index_run_id, updated_at)
                VALUES (%s, %s, %s, NOW())
                ON CONFLICT (tenant_id, domain_id) DO UPDATE
                SET index_run_id = EXCLUDED.index_run_id, updated_at = NOW()
                """,
                (tenant_id, domain_id, run_id),
            )
        conn.commit()


def ensure_active_index_run(tenant_id: str, domain_id: str) -> str:
    with _connect() as conn:
        with conn.cursor() as cur:
            cur.execute(
                "SELECT index_run_id FROM index_run_active WHERE tenant_id = %s AND domain_id = %s",
                (tenant_id, domain_id),
            )
            row = cur.fetchone()
            if row:
                return str(row[0])
    run_id = create_index_run(tenant_id, domain_id)
    activate_index_run(tenant_id, domain_id, run_id)
    return run_id


def active_index_run_id(tenant_id: str, domain_id: str) -> str | None:
    with _connect() as conn:
        with conn.cursor() as cur:
            cur.execute(
                "SELECT index_run_id FROM index_run_active WHERE tenant_id = %s AND domain_id = %s",
                (tenant_id, domain_id),
            )
            row = cur.fetchone()
    return str(row[0]) if row else None


def upsert_index_document_state(
    run_id: str,
    document_id: str,
    indexed_version: int,
    content_sha256: str,
    chunk_count: int,
) -> None:
    with _connect() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO index_document_state
                    (index_run_id, document_id, indexed_version, content_sha256, chunk_count, indexed_at)
                VALUES (%s, %s, %s, %s, %s, NOW())
                ON CONFLICT (index_run_id, document_id) DO UPDATE SET
                    indexed_version = EXCLUDED.indexed_version,
                    content_sha256 = EXCLUDED.content_sha256,
                    chunk_count = EXCLUDED.chunk_count,
                    indexed_at = NOW()
                """,
                (run_id, document_id, indexed_version, content_sha256, chunk_count),
            )
        conn.commit()


def collection_suffix(tenant_id: str, domain_id: str) -> str:
    """Namespace suffix for disposable vector collections."""
    from rag.kb.index_collections import run_suffix

    run_id = active_index_run_id(tenant_id, domain_id)
    if not run_id:
        return ""
    return run_suffix(run_id)


def resolve_write_run_id(tenant_id: str, domain_id: str, explicit_run_id: str | None = None) -> str:
    """Run id for ingest writes: explicit building run or active (creating one if needed)."""
    if explicit_run_id:
        return explicit_run_id
    return ensure_active_index_run(tenant_id, domain_id)


def resolve_read_run_id(tenant_id: str, domain_id: str) -> str | None:
    """Active index run for retrieval; None → legacy flat index."""
    return active_index_run_id(tenant_id, domain_id)


def list_retired_run_ids(tenant_id: str, domain_id: str, *, keep_last: int = 1) -> list[str]:
    """Retired run ids eligible for GC (newest `keep_last` retired runs are kept)."""
    keep = max(0, int(keep_last))
    with _connect() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT id FROM index_runs
                WHERE status = 'retired'
                  AND tenant_id IS NOT DISTINCT FROM NULLIF(%s, '')::text
                  AND domain_id IS NOT DISTINCT FROM NULLIF(%s, '')::text
                ORDER BY activated_at DESC NULLS LAST, created_at DESC
                OFFSET %s
                """,
                (tenant_id, domain_id, keep),
            )
            rows = cur.fetchall()
    return [str(row[0]) for row in rows]
