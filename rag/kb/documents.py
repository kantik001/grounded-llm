"""Postgres KB document registry (source-of-truth metadata)."""

from __future__ import annotations

import hashlib
import json
import os
import uuid
from contextlib import contextmanager
from dataclasses import dataclass
from typing import Any

from rag.storage.blob_store import get_blob_store


def _database_url() -> str:
    url = (os.environ.get("DATABASE_URL") or "").strip()
    if not url:
        raise RuntimeError("DATABASE_URL is required for KB document registry")
    return url


def _psycopg_dsn(connection: str) -> str:
    return connection.replace("postgresql+psycopg://", "postgresql://", 1)


@contextmanager
def _connect():
    import psycopg

    with psycopg.connect(_psycopg_dsn(_database_url())) as conn:
        yield conn


def content_sha256(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


@dataclass
class KBDocument:
    id: str
    tenant_id: str
    domain_id: str
    logical_key: str
    title: str = ""
    mime_type: str = "application/octet-stream"
    status: str = "active"
    current_version: int = 0
    storage_key: str = ""
    content_sha256: str = ""
    version_id: str = ""


@dataclass
class DocumentTarget:
    document_id: str
    version_id: str
    tenant_id: str
    domain_id: str
    logical_key: str
    content_sha256: str
    storage_key: str
    mime_type: str = "application/octet-stream"
    local_path: str = ""
    current_version: int = 0


def upsert_document(
    *,
    tenant_id: str,
    domain_id: str,
    logical_key: str,
    data: bytes,
    mime_type: str = "application/octet-stream",
    title: str = "",
    source: str = "upload",
    source_ref: dict[str, Any] | None = None,
    created_by: str = "",
) -> KBDocument:
    """Create/update document metadata and store blob (Postgres + object storage)."""
    sha = content_sha256(data)
    blob = get_blob_store()
    doc_id = ""
    version = 0

    with _connect() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT id, current_version FROM kb_documents
                WHERE tenant_id = %s AND domain_id = %s AND logical_key = %s
                """,
                (tenant_id, domain_id, logical_key),
            )
            row = cur.fetchone()
            if row:
                doc_id = str(row[0])
                version = int(row[1] or 0)
            else:
                doc_id = str(uuid.uuid4())

            next_version = version + 1
            ext = logical_key.rsplit(".", 1)[-1] if "." in logical_key else ""
            storage_key = blob.build_key(tenant_id, domain_id, doc_id, next_version, sha, ext)
            blob.put(storage_key, data, content_type=mime_type)

            if version == 0:
                cur.execute(
                    """
                    INSERT INTO kb_documents
                        (id, tenant_id, domain_id, logical_key, title, mime_type, status, current_version)
                    VALUES (%s, %s, %s, %s, %s, %s, 'active', 0)
                    """,
                    (doc_id, tenant_id, domain_id, logical_key, title or logical_key, mime_type),
                )

            version_id = str(uuid.uuid4())
            cur.execute(
                """
                INSERT INTO kb_document_versions
                    (id, document_id, version, storage_key, content_sha256, size_bytes, source, source_ref, created_by)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s::jsonb, %s)
                """,
                (
                    version_id,
                    doc_id,
                    next_version,
                    storage_key,
                    sha,
                    len(data),
                    source,
                    json.dumps(source_ref or {}),
                    created_by,
                ),
            )
            cur.execute(
                """
                UPDATE kb_documents
                SET current_version = %s, title = COALESCE(NULLIF(%s, ''), title),
                    mime_type = COALESCE(NULLIF(%s, ''), mime_type),
                    status = 'active', updated_at = NOW()
                WHERE id = %s
                """,
                (next_version, title, mime_type, doc_id),
            )
            cur.execute(
                """
                INSERT INTO kb_document_acl (document_id, principal_type, principal_id, permission)
                VALUES (%s, 'tenant', %s, 'read')
                ON CONFLICT (document_id, principal_type, principal_id) DO NOTHING
                """,
                (doc_id, tenant_id),
            )
        conn.commit()

    return KBDocument(
        id=doc_id,
        tenant_id=tenant_id,
        domain_id=domain_id,
        logical_key=logical_key,
        title=title or logical_key,
        mime_type=mime_type,
        current_version=next_version,
        storage_key=storage_key,
        content_sha256=sha,
        version_id=version_id,
    )


def list_active_documents(tenant_id: str, domain_id: str, *, logical_keys: list[str] | None = None) -> list[DocumentTarget]:
    clauses = ["tenant_id = %s", "domain_id = %s", "status = 'active'"]
    params: list[Any] = [tenant_id, domain_id]
    if logical_keys:
        clauses.append("logical_key = ANY(%s)")
        params.append(logical_keys)

    sql = f"""
        SELECT d.id, d.tenant_id, d.domain_id, d.logical_key, d.title, d.mime_type,
               d.status, d.current_version, v.id, v.storage_key, v.content_sha256
        FROM kb_documents d
        JOIN kb_document_versions v
          ON v.document_id = d.id AND v.version = d.current_version
        WHERE {' AND '.join(clauses)}
        ORDER BY d.logical_key
    """
    out: list[DocumentTarget] = []
    with _connect() as conn:
        with conn.cursor() as cur:
            cur.execute(sql, params)
            for row in cur.fetchall():
                out.append(
                    DocumentTarget(
                        document_id=str(row[0]),
                        version_id=str(row[8]),
                        tenant_id=row[1],
                        domain_id=row[2],
                        logical_key=row[3],
                        content_sha256=row[10] or "",
                        storage_key=row[9] or "",
                        mime_type=row[5] or "application/octet-stream",
                        current_version=int(row[7] or 0),
                    )
                )
    return out


def materialize_to_temp(target: DocumentTarget) -> str:
    """Download blob to a temp file for parsers (returns path)."""
    import tempfile

    blob = get_blob_store()
    data = blob.get(target.storage_key)
    ext = ""
    if "." in target.logical_key:
        ext = "." + target.logical_key.rsplit(".", 1)[-1]
    fd, path = tempfile.mkstemp(suffix=ext, prefix="kb_doc_")
    os.close(fd)
    with open(path, "wb") as fh:
        fh.write(data)
    return path


def list_all_active_documents() -> list[DocumentTarget]:
    """All active documents across tenants (full reindex / manifest scan)."""
    sql = """
        SELECT d.id, d.tenant_id, d.domain_id, d.logical_key, d.mime_type,
               d.current_version, v.id, v.storage_key, v.content_sha256
        FROM kb_documents d
        JOIN kb_document_versions v
          ON v.document_id = d.id AND v.version = d.current_version
        WHERE d.status = 'active'
        ORDER BY d.tenant_id, d.domain_id, d.logical_key
    """
    out: list[DocumentTarget] = []
    with _connect() as conn:
        with conn.cursor() as cur:
            cur.execute(sql)
            for row in cur.fetchall():
                out.append(
                    DocumentTarget(
                        document_id=str(row[0]),
                        version_id=str(row[6]),
                        tenant_id=row[1],
                        domain_id=row[2],
                        logical_key=row[3],
                        content_sha256=row[8] or "",
                        storage_key=row[7] or "",
                        mime_type=row[4] or "application/octet-stream",
                        current_version=int(row[5] or 0),
                    )
                )
    return out


def scan_registry_documents() -> dict[str, dict]:
    """Current KB state for Chroma manifest: {tenant/domain/file: metadata}."""
    from rag.document_loaders import is_supported_filename

    state: dict[str, dict] = {}
    for target in list_all_active_documents():
        if not is_supported_filename(target.logical_key):
            continue
        key = f"{target.tenant_id}/{target.domain_id}/{target.logical_key}"
        state[key] = {
            "sha1": target.content_sha256,
            "sha256": target.content_sha256,
            "storage_key": target.storage_key,
            "document_id": target.document_id,
            "tenant": target.tenant_id,
            "domain": target.domain_id,
            "filename": target.logical_key,
        }
    return state


def discover_document_targets(
    tenant_id: str,
    domain_id: str,
    files: list[str] | None = None,
) -> list[DocumentTarget]:
    """Resolve ingest targets from Postgres registry."""
    keys = [f for f in (files or []) if f]
    return list_active_documents(tenant_id, domain_id, logical_keys=keys or None)


def mark_deleted(tenant_id: str, domain_id: str, logical_key: str) -> None:
    with _connect() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                UPDATE kb_documents SET status = 'deleted', updated_at = NOW()
                WHERE tenant_id = %s AND domain_id = %s AND logical_key = %s
                """,
                (tenant_id, domain_id, logical_key),
            )
        conn.commit()


def allowed_document_ids(
    tenant_id: str,
    principals: list[tuple[str, str]] | None = None,
) -> list[str]:
    """Return document ids readable by principals (default: whole tenant)."""
    if not principals:
        principals = [("tenant", tenant_id)]
    seen: set[str] = set()
    out: list[str] = []
    with _connect() as conn:
        with conn.cursor() as cur:
            for ptype, pid in principals:
                cur.execute(
                    """
                    SELECT DISTINCT d.id
                    FROM kb_documents d
                    JOIN kb_document_acl a ON a.document_id = d.id
                    WHERE d.tenant_id = %s AND d.status = 'active'
                      AND a.principal_type = %s AND a.principal_id = %s
                      AND a.permission IN ('read', 'admin')
                    """,
                    (tenant_id, ptype, pid),
                )
                for (doc_id,) in cur.fetchall():
                    sid = str(doc_id)
                    if sid not in seen:
                        seen.add(sid)
                        out.append(sid)
    return out
