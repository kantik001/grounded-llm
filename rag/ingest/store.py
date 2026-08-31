"""Postgres persistence for ingest jobs and tasks."""

from __future__ import annotations

import json
import os
from contextlib import contextmanager
from dataclasses import dataclass, field
from datetime import datetime, timezone
from typing import Any

from rag.ingest.models import (
    STAGE_EMBED,
    STAGE_FINALIZE,
    STAGE_PARSE,
    TERMINAL_JOB_STATUSES,
    IngestTaskStatus,
)
from rag.vector_backend.pgvector_backend import psycopg_dsn


def _database_url() -> str:
    url = (os.environ.get("DATABASE_URL") or "").strip()
    if not url:
        raise RuntimeError("DATABASE_URL is required for ingest jobs")
    return url


@contextmanager
def _connect():
    import psycopg

    with psycopg.connect(psycopg_dsn(_database_url())) as conn:
        yield conn


def _utc_now() -> datetime:
    return datetime.now(timezone.utc)


@dataclass
class IngestJob:
    id: int
    status: str
    tenant_id: str
    domain_id: str
    source: str
    actor: str
    mode: str
    files: list[str] = field(default_factory=list)
    stats: dict[str, Any] = field(default_factory=dict)
    error_msg: str = ""
    attempt_count: int = 0
    started_at: str = ""
    finished_at: str = ""
    created_at: str = ""


@dataclass
class IngestTask:
    id: int
    job_id: int
    stage: str
    file_key: str
    payload: dict[str, Any]
    status: str
    attempts: int
    max_attempts: int
    error_msg: str = ""


def _fmt_ts(value: datetime | None) -> str:
    if value is None:
        return ""
    return value.astimezone(timezone.utc).isoformat()


def _row_to_job(row: tuple) -> IngestJob:
    files_raw = row[7]
    stats_raw = row[8]
    files = files_raw if isinstance(files_raw, list) else json.loads(files_raw or "[]")
    stats = stats_raw if isinstance(stats_raw, dict) else json.loads(stats_raw or "{}")
    return IngestJob(
        id=row[0],
        status=row[1],
        tenant_id=row[2],
        domain_id=row[3],
        source=row[4] or "admin",
        actor=row[5] or "",
        mode=row[6] or "incremental",
        files=[str(x) for x in files],
        stats=stats if isinstance(stats, dict) else {},
        error_msg=row[9] or "",
        attempt_count=int(row[10] or 0),
        started_at=_fmt_ts(row[11]),
        finished_at=_fmt_ts(row[12]),
        created_at=_fmt_ts(row[13]),
    )


def get_job(job_id: int) -> IngestJob | None:
    with _connect() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT id, status, tenant_id, domain_id, source, actor, mode,
                       files, stats, error_msg, attempt_count, started_at, finished_at, created_at
                FROM ingest_jobs WHERE id = %s
                """,
                (job_id,),
            )
            row = cur.fetchone()
    if not row:
        return None
    return _row_to_job(row)


def update_job_status(
    job_id: int,
    status: str,
    *,
    stats: dict[str, Any] | None = None,
    error_msg: str = "",
    mark_started: bool = False,
    mark_finished: bool = False,
) -> None:
    sets = ["status = %s"]
    params: list[Any] = [status]
    if stats is not None:
        sets.append("stats = %s::jsonb")
        params.append(json.dumps(stats))
    if error_msg:
        sets.append("error_msg = %s")
        params.append(error_msg)
    if mark_started:
        sets.append("started_at = COALESCE(started_at, NOW())")
    if mark_finished:
        sets.append("finished_at = NOW()")
    params.append(job_id)
    sql = f"UPDATE ingest_jobs SET {', '.join(sets)} WHERE id = %s"
    with _connect() as conn:
        with conn.cursor() as cur:
            cur.execute(sql, params)
        conn.commit()


def merge_job_stats(job_id: int, patch: dict[str, Any]) -> dict[str, Any]:
    job = get_job(job_id)
    if job is None:
        return patch
    merged = {**job.stats, **patch}
    update_job_status(job_id, job.status, stats=merged)
    return merged


def create_task(
    job_id: int,
    stage: str,
    *,
    file_key: str = "",
    payload: dict[str, Any] | None = None,
) -> int:
    with _connect() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO ingest_tasks (job_id, stage, file_key, payload, status)
                VALUES (%s, %s, %s, %s::jsonb, %s)
                RETURNING id
                """,
                (
                    job_id,
                    stage,
                    file_key or None,
                    json.dumps(payload or {}),
                    IngestTaskStatus.PENDING,
                ),
            )
            task_id = int(cur.fetchone()[0])
        conn.commit()
    return task_id


def claim_task(stage: str, *, lease_sec: int = 300) -> IngestTask | None:
    with _connect() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT id, job_id, stage, file_key, payload, status, attempts, max_attempts, error_msg
                FROM ingest_tasks
                WHERE stage = %s AND status = %s
                ORDER BY created_at ASC
                FOR UPDATE SKIP LOCKED
                LIMIT 1
                """,
                (stage, IngestTaskStatus.PENDING),
            )
            row = cur.fetchone()
            if not row:
                return None
            task_id = row[0]
            cur.execute(
                """
                UPDATE ingest_tasks
                SET status = %s,
                    attempts = attempts + 1,
                    lease_until = NOW() + (%s || ' seconds')::interval,
                    updated_at = NOW()
                WHERE id = %s
                RETURNING id, job_id, stage, file_key, payload, status, attempts, max_attempts, error_msg
                """,
                (IngestTaskStatus.PROCESSING, str(lease_sec), task_id),
            )
            row = cur.fetchone()
        conn.commit()
    payload = row[4]
    if not isinstance(payload, dict):
        payload = json.loads(payload or "{}")
    return IngestTask(
        id=row[0],
        job_id=row[1],
        stage=row[2],
        file_key=row[3] or "",
        payload=payload,
        status=row[5],
        attempts=row[6],
        max_attempts=row[7],
        error_msg=row[8] or "",
    )


def try_claim_task(task_id: int, *, lease_sec: int = 300) -> IngestTask | None:
    with _connect() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                UPDATE ingest_tasks
                SET status = %s,
                    attempts = attempts + 1,
                    lease_until = NOW() + (%s || ' seconds')::interval,
                    updated_at = NOW()
                WHERE id = %s AND status = %s
                RETURNING id, job_id, stage, file_key, payload, status, attempts, max_attempts, error_msg
                """,
                (IngestTaskStatus.PROCESSING, str(lease_sec), task_id, IngestTaskStatus.PENDING),
            )
            row = cur.fetchone()
        conn.commit()
    if not row:
        return None
    payload = row[4]
    if not isinstance(payload, dict):
        payload = json.loads(payload or "{}")
    return IngestTask(
        id=row[0],
        job_id=row[1],
        stage=row[2],
        file_key=row[3] or "",
        payload=payload,
        status=row[5],
        attempts=row[6],
        max_attempts=row[7],
        error_msg=row[8] or "",
    )


def finish_task(task_id: int, *, status: str = IngestTaskStatus.DONE, error_msg: str = "") -> None:
    with _connect() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                UPDATE ingest_tasks
                SET status = %s, error_msg = %s, updated_at = NOW(), lease_until = NULL
                WHERE id = %s
                """,
                (status, error_msg or None, task_id),
            )
        conn.commit()


def count_tasks(job_id: int, *, stage: str | None = None, status: str | None = None) -> int:
    clauses = ["job_id = %s"]
    params: list[Any] = [job_id]
    if stage:
        clauses.append("stage = %s")
        params.append(stage)
    if status:
        clauses.append("status = %s")
        params.append(status)
    sql = f"SELECT COUNT(*) FROM ingest_tasks WHERE {' AND '.join(clauses)}"
    with _connect() as conn:
        with conn.cursor() as cur:
            cur.execute(sql, params)
            return int(cur.fetchone()[0])


def list_tasks(job_id: int) -> list[dict[str, Any]]:
    with _connect() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT id, stage, file_key, status, attempts, max_attempts, error_msg, updated_at
                FROM ingest_tasks WHERE job_id = %s ORDER BY id
                """,
                (job_id,),
            )
            rows = cur.fetchall()
    out: list[dict[str, Any]] = []
    for row in rows:
        out.append(
            {
                "id": row[0],
                "stage": row[1],
                "file_key": row[2],
                "status": row[3],
                "attempts": row[4],
                "max_attempts": row[5],
                "error_msg": row[6] or "",
                "updated_at": _fmt_ts(row[7]),
            }
        )
    return out


def job_status_payload(job_id: int) -> dict[str, Any] | None:
    job = get_job(job_id)
    if job is None:
        return None
    return {
        "job": {
            "id": job.id,
            "status": job.status,
            "tenant_id": job.tenant_id,
            "domain_id": job.domain_id,
            "source": job.source,
            "mode": job.mode,
            "files": job.files,
            "stats": job.stats,
            "error_msg": job.error_msg,
            "started_at": job.started_at,
            "finished_at": job.finished_at,
            "created_at": job.created_at,
        },
        "tasks": list_tasks(job_id),
        "done": job.status in TERMINAL_JOB_STATUSES,
    }


def get_task(task_id: int) -> IngestTask | None:
    with _connect() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT id, job_id, stage, file_key, payload, status, attempts, max_attempts, error_msg
                FROM ingest_tasks WHERE id = %s
                """,
                (task_id,),
            )
            row = cur.fetchone()
    if not row:
        return None
    payload = row[4]
    if not isinstance(payload, dict):
        payload = json.loads(payload or "{}")
    return IngestTask(
        id=row[0],
        job_id=row[1],
        stage=row[2],
        file_key=row[3] or "",
        payload=payload,
        status=row[5],
        attempts=row[6],
        max_attempts=row[7],
        error_msg=row[8] or "",
    )


def reset_failed_task(task_id: int) -> None:
    with _connect() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                UPDATE ingest_tasks
                SET status = %s, error_msg = NULL, lease_until = NULL, updated_at = NOW()
                WHERE id = %s AND status = %s
                """,
                (IngestTaskStatus.PENDING, task_id, IngestTaskStatus.FAILED),
            )
        conn.commit()


def requeue_stale_tasks(*, lease_sec: int = 300) -> int:
    with _connect() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                UPDATE ingest_tasks
                SET status = %s, lease_until = NULL, updated_at = NOW()
                WHERE status = %s
                  AND lease_until IS NOT NULL
                  AND lease_until < NOW()
                """,
                (IngestTaskStatus.PENDING, IngestTaskStatus.PROCESSING),
            )
            n = cur.rowcount
        conn.commit()
    return int(n)


__all__ = [
    "STAGE_EMBED",
    "STAGE_FINALIZE",
    "STAGE_PARSE",
    "IngestJob",
    "IngestTask",
    "claim_task",
    "try_claim_task",
    "count_tasks",
    "create_task",
    "finish_task",
    "get_task",
    "get_job",
    "job_status_payload",
    "merge_job_stats",
    "requeue_stale_tasks",
    "reset_failed_task",
    "update_job_status",
]
