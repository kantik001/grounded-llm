"""KB ingest outbox: registry upserts → async ingest enqueue (KB_AUTO_INGEST=1)."""

from __future__ import annotations

import os
from dataclasses import dataclass
from typing import Any

from rag.kb.db import connect as _connect


def auto_ingest_enabled() -> bool:
    return os.environ.get("KB_AUTO_INGEST", "0").strip().lower() in ("1", "true", "yes")


def enqueue_outbox_tx(
    cur,
    *,
    tenant_id: str,
    domain_id: str,
    document_id: str,
    version_id: str,
    logical_key: str,
    content_sha256: str,
    source: str = "upload",
) -> None:
    """Insert outbox row in the caller's open transaction."""
    if not auto_ingest_enabled():
        return
    cur.execute(
        """
        INSERT INTO kb_ingest_outbox
            (tenant_id, domain_id, document_id, version_id, logical_key, content_sha256, source, status)
        VALUES (%s, %s, %s, %s, %s, %s, %s, 'pending')
        """,
        (tenant_id, domain_id, document_id, version_id, logical_key, content_sha256, source or "upload"),
    )


@dataclass
class OutboxFlushResult:
    tenant_id: str
    domain_id: str
    job_id: int | None
    already_running: bool
    flushed: int
    error: str = ""


def _server_base_url() -> str:
    base = (os.environ.get("GROUNDED_SERVER_URL") or os.environ.get("SERVER_URL") or "http://127.0.0.1:8080").strip()
    return base.rstrip("/")


def _admin_auth_headers() -> dict[str, str]:
    headers: dict[str, str] = {"Content-Type": "application/json"}
    secret = (os.environ.get("ADMIN_SECRET") or "").strip()
    if secret:
        headers["X-Admin-Secret"] = secret
    return headers


def _admin_auth() -> tuple[str, str] | None:
    user = (os.environ.get("ADMIN_USER") or "").strip()
    password = (os.environ.get("ADMIN_PASSWORD") or "").strip()
    if user and password:
        return user, password
    return None


def trigger_ingest_http(
    *,
    tenant_id: str,
    domain_id: str,
    files: list[str] | None = None,
    source: str = "outbox",
    sync: bool = False,
) -> dict[str, Any]:
    """Call Go admin ingest API (consumer stage 1)."""
    import requests

    url = f"{_server_base_url()}/api/admin/ingest"
    params: dict[str, str] = {"domain_id": domain_id}
    if tenant_id and tenant_id != "default":
        params["tenant_id"] = tenant_id
    headers = _admin_auth_headers()
    body = {
        "files": files or [],
        "mode": "incremental",
        "source": source,
        "sync": sync,
    }
    auth = _admin_auth()
    resp = requests.post(url, params=params, headers=headers, json=body, auth=auth, timeout=120)
    resp.raise_for_status()
    return resp.json()


def _claim_pending(tenant_id: str, domain_id: str) -> tuple[list[int], list[str], str]:
    ids: list[int] = []
    keys: list[str] = []
    source = "outbox"
    seen: set[str] = set()
    with _connect() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT id, logical_key, source
                FROM kb_ingest_outbox
                WHERE status = 'pending' AND tenant_id = %s AND domain_id = %s
                ORDER BY created_at
                FOR UPDATE SKIP LOCKED
                """,
                (tenant_id, domain_id),
            )
            rows = cur.fetchall()
            if not rows:
                return [], [], source
            for row_id, logical_key, row_source in rows:
                ids.append(int(row_id))
                if logical_key and logical_key not in seen:
                    seen.add(logical_key)
                    keys.append(logical_key)
                if row_source:
                    source = row_source
            cur.execute(
                "UPDATE kb_ingest_outbox SET status = 'processing' WHERE id = ANY(%s)",
                (ids,),
            )
        conn.commit()
    return ids, keys, source


def _finish_outbox(ids: list[int], job_id: int | None, error: str = "") -> None:
    if not ids:
        return
    status = "failed" if error else "done"
    with _connect() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                UPDATE kb_ingest_outbox
                SET status = %s,
                    ingest_job_id = COALESCE(%s, ingest_job_id),
                    error_msg = NULLIF(%s, ''),
                    processed_at = NOW()
                WHERE id = ANY(%s)
                """,
                (status, job_id, error, ids),
            )
        conn.commit()


def flush_outbox(
    *,
    tenant_id: str,
    domain_id: str,
    sync: bool = False,
) -> OutboxFlushResult:
    """Claim pending outbox rows and trigger Go ingest API."""
    if not auto_ingest_enabled():
        return OutboxFlushResult(tenant_id, domain_id, None, False, 0)

    ids, files, source = _claim_pending(tenant_id, domain_id)
    if not ids:
        return OutboxFlushResult(tenant_id, domain_id, None, False, 0)

    try:
        payload = trigger_ingest_http(
            tenant_id=tenant_id,
            domain_id=domain_id,
            files=files,
            source=source,
            sync=sync,
        )
        job_id = payload.get("job_id")
        already = bool(payload.get("already_running"))
        _finish_outbox(ids, int(job_id) if job_id else None)
        return OutboxFlushResult(tenant_id, domain_id, int(job_id) if job_id else None, already, len(ids))
    except Exception as exc:
        _finish_outbox(ids, None, str(exc))
        return OutboxFlushResult(tenant_id, domain_id, None, False, len(ids), error=str(exc))
