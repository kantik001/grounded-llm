"""Ingest worker loops (sync drain + Redis consumers)."""

from __future__ import annotations

import os
import time

from rag.ingest import metrics
from rag.ingest.models import STAGE_EMBED, STAGE_FINALIZE, STAGE_PARSE, IngestTaskStatus
from rag.ingest import pipeline
from rag.ingest import queue as ingest_queue
from rag.ingest import store as ingest_store


def drain_job_sync(job_id: int) -> dict:
    """Process all pending tasks for a job in-process (week 1 fake-async)."""
    while True:
        ingest_store.requeue_stale_tasks()
        task = _next_pending_task(job_id)
        if task is None:
            break
        pipeline.process_task(task.stage, task.id)
        job = ingest_store.get_job(job_id)
        if job and job.status in ("succeeded", "failed", "partial"):
            break
    return ingest_store.job_status_payload(job_id) or {}


def _next_pending_task(job_id: int) -> ingest_store.IngestTask | None:
    for stage in (STAGE_PARSE, STAGE_EMBED, STAGE_FINALIZE):
        with ingest_store._connect() as conn:
            with conn.cursor() as cur:
                cur.execute(
                    """
                    SELECT id FROM ingest_tasks
                    WHERE job_id = %s AND stage = %s AND status = %s
                    ORDER BY id LIMIT 1
                    """,
                    (job_id, stage, IngestTaskStatus.PENDING),
                )
                row = cur.fetchone()
        if row:
            return ingest_store.get_task(int(row[0]))
    return None


def run_worker(stage: str, *, once: bool = False) -> None:
    idle_rounds = 0
    while True:
        ingest_store.requeue_stale_tasks()
        payload = ingest_queue.dequeue(stage, timeout_sec=3)
        if payload is None:
            idle_rounds += 1
            if once or idle_rounds > 40:
                return
            continue
        idle_rounds = 0
        task_id = int(payload.get("task_id") or 0)
        if task_id <= 0:
            metrics.inc("tasks.invalid_payload")
            continue
        metrics.inc(f"worker.{stage}.received")
        pipeline.process_task(stage, task_id)
        if once:
            return


def run_worker_poll(stage: str, *, once: bool = False) -> None:
    """Postgres SKIP LOCKED fallback when Redis is unavailable."""
    while True:
        ingest_store.requeue_stale_tasks()
        task = ingest_store.claim_task(stage)
        if task is None:
            if once:
                return
            time.sleep(1)
            continue
        metrics.inc(f"worker.{stage}.claimed")
        pipeline.process_task(stage, task.stage)
        if once:
            return


def run_multi_worker(*, once: bool = False) -> None:
    """Poll parse → embed → finalize queues in one process."""
    idle_rounds = 0
    stages = (STAGE_PARSE, STAGE_EMBED, STAGE_FINALIZE)
    while True:
        ingest_store.requeue_stale_tasks()
        handled = False
        for stage in stages:
            payload = ingest_queue.dequeue(stage, timeout_sec=1)
            if payload is None:
                continue
            handled = True
            idle_rounds = 0
            task_id = int(payload.get("task_id") or 0)
            if task_id <= 0:
                metrics.inc("tasks.invalid_payload")
                continue
            metrics.inc(f"worker.{stage}.received")
            pipeline.process_task(stage, task_id)
            if once:
                return
        if not handled:
            idle_rounds += 1
            if once or idle_rounds > 60:
                return


def main() -> None:
    stage = (os.environ.get("INGEST_WORKER_STAGE") or "all").strip().lower()
    once = (os.environ.get("INGEST_WORKER_ONCE") or "").strip().lower() in ("1", "true", "yes")
    use_redis = (os.environ.get("INGEST_USE_REDIS") or "1").strip().lower() not in ("0", "false", "no")

    if stage == "all" and use_redis:
        run_multi_worker(once=once)
        return

    stages = [STAGE_PARSE, STAGE_EMBED, STAGE_FINALIZE] if stage == "all" else [stage]
    for st in stages:
        if use_redis:
            run_worker(st, once=once)
        else:
            run_worker_poll(st, once=once)


if __name__ == "__main__":
    main()
