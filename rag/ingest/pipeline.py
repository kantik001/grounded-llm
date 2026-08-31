"""Ingest pipeline stages: discover → parse → embed → finalize."""

from __future__ import annotations

import json
import os
from dataclasses import dataclass
from typing import Any

from langchain_core.documents import Document

from rag.indexing import split_file_documents
from rag.ingest import metrics
from rag.ingest import queue as ingest_queue
from rag.ingest import store as ingest_store
from rag.ingest.models import (
    STAGE_EMBED,
    STAGE_FINALIZE,
    STAGE_PARSE,
    IngestJobStatus,
    IngestTaskStatus,
)
from rag.kb.documents import DocumentTarget, discover_document_targets, materialize_to_temp, scan_registry_documents
from rag.kb.index_runs import ensure_active_index_run, upsert_index_document_state
from rag.sparse_index import ensure_sparse_index
from rag.vector_backend import get_vector_backend
from rag.vector_backend.chroma_backend import ChromaBackend


def staging_root() -> str:
    root = os.environ.get("INGEST_STAGING_DIR", "ingest_staging").strip() or "ingest_staging"
    if not os.path.isabs(root):
        base = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
        root = os.path.join(base, root)
    return root


def staging_path(file_key: str) -> str:
    safe = file_key.replace("/", "__")
    return os.path.join(staging_root(), f"{safe}.chunks.jsonl")


def async_enabled() -> bool:
    raw = (os.environ.get("INGEST_ASYNC") or "1").strip().lower()
    return raw not in ("0", "false", "no")


@dataclass
class FileTarget:
    file_key: str
    path: str
    tenant_id: str
    domain_id: str
    filename: str


def discover_targets(job: ingest_store.IngestJob) -> list[DocumentTarget]:
    """Resolve ingest targets from Postgres registry."""
    explicit = [f for f in job.files if f]
    return discover_document_targets(job.tenant_id, job.domain_id, explicit or None)


def _target_file_key(target: DocumentTarget) -> str:
    return f"{target.tenant_id}/{target.domain_id}/{target.logical_key}"


def _target_to_file(target: DocumentTarget) -> FileTarget:
    path = target.local_path
    if not path and target.storage_key:
        path = materialize_to_temp(target)
    return FileTarget(
        file_key=_target_file_key(target),
        path=path,
        tenant_id=target.tenant_id,
        domain_id=target.domain_id,
        filename=target.logical_key,
    )


def _resolve_parse_path(task: ingest_store.IngestTask) -> tuple[str, bool]:
    """Return (path, is_temp)."""
    payload = task.payload
    path = payload.get("path") or ""
    if path and os.path.isfile(path):
        return path, False
    storage_key = payload.get("storage_key") or ""
    if storage_key:
        target = DocumentTarget(
            document_id=payload.get("document_id") or "",
            version_id=payload.get("version_id") or "",
            tenant_id=payload.get("tenant_id") or "",
            domain_id=payload.get("domain_id") or "",
            logical_key=payload.get("filename") or payload.get("logical_key") or "",
            content_sha256=payload.get("content_sha256") or "",
            storage_key=storage_key,
        )
        return materialize_to_temp(target), True
    raise ValueError("parse task missing path or storage_key")


def _enrich_chunks(docs: list[Document], payload: dict[str, Any]) -> None:
    doc_id = payload.get("document_id") or ""
    version_id = payload.get("version_id") or ""
    if not doc_id:
        return
    for doc in docs:
        meta = doc.metadata or {}
        meta["document_id"] = doc_id
        if version_id:
            meta["document_version_id"] = version_id
        doc.metadata = meta


def _write_staging(file_key: str, docs: list[Document]) -> str:
    path = staging_path(file_key)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w", encoding="utf-8") as fh:
        for doc in docs:
            row = {"page_content": doc.page_content, "metadata": doc.metadata}
            fh.write(json.dumps(row, ensure_ascii=False) + "\n")
    return path


def _read_staging(file_key: str) -> list[Document]:
    path = staging_path(file_key)
    docs: list[Document] = []
    with open(path, encoding="utf-8") as fh:
        for line in fh:
            line = line.strip()
            if not line:
                continue
            row = json.loads(line)
            docs.append(Document(page_content=row["page_content"], metadata=row.get("metadata") or {}))
    return docs


def _upsert_chroma_file(target: FileTarget) -> int:
    backend = get_vector_backend()
    if isinstance(backend, ChromaBackend):
        return backend.upsert_kb_file(
            target.tenant_id,
            target.domain_id,
            target.path,
            filename=target.filename,
        )
    chunks = split_file_documents(target.domain_id, target.path, tenant_id=target.tenant_id)
    backend.load()
    store = getattr(backend, "_store", None)
    if store is not None and chunks:
        store.add_documents(chunks)
    return len(chunks)


def run_parse(task: ingest_store.IngestTask) -> dict[str, Any]:
    payload = task.payload
    domain_id = payload.get("domain_id") or ""
    tenant_id = payload.get("tenant_id") or ""
    if not domain_id:
        raise ValueError("parse task missing domain_id")
    path, is_temp = _resolve_parse_path(task)
    try:
        with metrics.timer("parse_duration"):
            docs = split_file_documents(domain_id, path, tenant_id=tenant_id)
            _enrich_chunks(docs, payload)
        staging = _write_staging(task.file_key, docs)
        return {"staging_path": staging, "chunks": len(docs)}
    finally:
        if is_temp and os.path.isfile(path):
            try:
                os.remove(path)
            except OSError:
                pass


def run_embed(task: ingest_store.IngestTask) -> dict[str, Any]:
    payload = task.payload
    path = payload.get("path") or ""
    if not path and payload.get("storage_key"):
        path = _resolve_parse_path(task)[0]
    target = FileTarget(
        file_key=task.file_key,
        path=path,
        tenant_id=payload.get("tenant_id") or "",
        domain_id=payload.get("domain_id") or "",
        filename=payload.get("filename") or payload.get("logical_key") or os.path.basename(path),
    )
    if not target.path:
        raise ValueError("embed task missing path")

    staging_file = staging_path(task.file_key)
    with metrics.timer("embed_duration"):
        if os.path.isfile(staging_file):
            docs = _read_staging(task.file_key)
            backend = get_vector_backend()
            backend.load()
            store = getattr(backend, "_store", None)
            if store is not None and docs:
                if isinstance(backend, ChromaBackend):
                    backend.delete_kb_file(target.tenant_id, target.domain_id, target.filename)
                    store.add_documents(docs)
                    backend.touch_manifest_entry(target.tenant_id, target.domain_id, target.path, target.filename)
                else:
                    store.add_documents(docs)
            embedded = len(docs)
        else:
            embedded = _upsert_chroma_file(target)

    doc_id = payload.get("document_id") or ""
    if doc_id:
        run_id = ensure_active_index_run(target.tenant_id, target.domain_id)
        upsert_index_document_state(
            run_id,
            doc_id,
            int(payload.get("document_version") or payload.get("version") or 0),
            payload.get("content_sha256") or "",
            embedded,
        )
    return {"embedded": embedded}


def run_finalize(job_id: int) -> dict[str, Any]:
    job = ingest_store.get_job(job_id)
    if job is None:
        raise ValueError(f"job {job_id} not found")
    with metrics.timer("finalize_duration"):
        ensure_active_index_run(job.tenant_id, job.domain_id)
        ensure_sparse_index(force_reindex=True)
        backend = get_vector_backend()
        if isinstance(backend, ChromaBackend):
            backend.sync_manifest(scan_registry_documents())
    stats = ingest_store.merge_job_stats(
        job_id,
        {
            "files_total": ingest_store.count_tasks(job_id, stage=STAGE_PARSE),
            "files_embedded": ingest_store.count_tasks(job_id, stage=STAGE_EMBED, status=IngestTaskStatus.DONE),
        },
    )
    return stats


def _enqueue_parse_tasks(
    job: ingest_store.IngestJob,
    targets: list[DocumentTarget],
    *,
    async_mode: bool,
) -> list[int]:
    task_ids: list[int] = []
    for target in targets:
        ft = _target_to_file(target)
        sha = target.content_sha256 or ""
        payload = {
            "path": ft.path,
            "tenant_id": target.tenant_id,
            "domain_id": target.domain_id,
            "filename": target.logical_key,
            "logical_key": target.logical_key,
            "sha1": sha,
            "document_id": target.document_id,
            "version_id": target.version_id,
            "document_version": target.current_version,
            "content_sha256": target.content_sha256,
            "storage_key": target.storage_key,
        }
        file_key = _target_file_key(target)
        task_id = ingest_store.create_task(
            job.id,
            STAGE_PARSE,
            file_key=file_key,
            payload=payload,
        )
        task_ids.append(task_id)
        if async_mode:
            ingest_queue.enqueue(
                STAGE_PARSE,
                {"task_id": task_id, "job_id": job.id, "file_key": file_key},
            )
    return task_ids


def _enqueue_embed_task(
    job_id: int,
    file_key: str,
    payload: dict[str, Any],
    parse_task_id: int,
    *,
    async_mode: bool,
) -> int:
    payload = {**payload, "parse_task_id": parse_task_id}
    task_id = ingest_store.create_task(job_id, STAGE_EMBED, file_key=file_key, payload=payload)
    if async_mode:
        ingest_queue.enqueue(
            STAGE_EMBED,
            {"task_id": task_id, "job_id": job_id, "file_key": file_key},
        )
    return task_id


def _maybe_enqueue_finalize(job_id: int, *, async_mode: bool) -> int | None:
    pending_embed = ingest_store.count_tasks(job_id, stage=STAGE_EMBED, status=IngestTaskStatus.PENDING)
    processing_embed = ingest_store.count_tasks(job_id, stage=STAGE_EMBED, status=IngestTaskStatus.PROCESSING)
    if pending_embed or processing_embed:
        return None
    failed_embed = ingest_store.count_tasks(job_id, stage=STAGE_EMBED, status=IngestTaskStatus.FAILED)
    dead_embed = ingest_store.count_tasks(job_id, stage=STAGE_EMBED, status=IngestTaskStatus.DEAD)
    if ingest_store.count_tasks(job_id, stage=STAGE_FINALIZE) > 0:
        return None
    ingest_store.update_job_status(job_id, IngestJobStatus.INDEXING, mark_started=True)
    task_id = ingest_store.create_task(job_id, STAGE_FINALIZE, payload={"job_id": job_id})
    if failed_embed or dead_embed:
        ingest_store.update_job_status(job_id, IngestJobStatus.PARTIAL)
    if async_mode:
        ingest_queue.enqueue(STAGE_FINALIZE, {"task_id": task_id, "job_id": job_id})
    return task_id


def start_job(job_id: int, *, sync: bool = False) -> dict[str, Any]:
    job = ingest_store.get_job(job_id)
    if job is None:
        raise ValueError(f"job {job_id} not found")
    if job.status not in (IngestJobStatus.QUEUED,):
        return ingest_store.job_status_payload(job_id) or {}

    if job.mode == "full":
        backend = get_vector_backend()
        if isinstance(backend, ChromaBackend):
            backend.load(force_reindex=True)
        else:
            backend.load(force_reindex=True)

    targets = discover_targets(job)
    if not targets:
        ingest_store.update_job_status(
            job_id,
            IngestJobStatus.FAILED,
            error_msg="No documents to ingest",
            mark_finished=True,
        )
        metrics.inc("jobs.failed")
        return ingest_store.job_status_payload(job_id) or {}

    async_mode = async_enabled() and not sync
    ingest_store.update_job_status(job_id, IngestJobStatus.PARSING, mark_started=True)
    metrics.inc("jobs.started")
    _enqueue_parse_tasks(job, targets, async_mode=async_mode)

    if sync or not async_mode:
        from rag.ingest.runner import drain_job_sync

        return drain_job_sync(job_id)
    return ingest_store.job_status_payload(job_id) or {}


def process_task(stage: str, task_id: int) -> None:
    task = ingest_store.try_claim_task(task_id) or ingest_store.get_task(task_id)
    if task is None:
        return
    if task.status == IngestTaskStatus.DONE:
        return
    if task.status == IngestTaskStatus.DEAD:
        return
    if task.stage != stage:
        raise ValueError(f"task {task_id} stage mismatch: {task.stage} != {stage}")

    job = ingest_store.get_job(task.job_id)
    if job is None:
        ingest_store.finish_task(task_id, status=IngestTaskStatus.DEAD, error_msg="job missing")
        return

    try:
        if stage == STAGE_PARSE:
            ingest_store.update_job_status(task.job_id, IngestJobStatus.PARSING)
            result = run_parse(task)
            ingest_store.finish_task(task_id)
            metrics.inc("tasks.parse.done")
            ingest_store.merge_job_stats(task.job_id, {"last_parse_chunks": result.get("chunks", 0)})
            embed_id = _enqueue_embed_task(
                task.job_id,
                task.file_key,
                dict(task.payload),
                task_id,
                async_mode=async_enabled(),
            )
            if not async_enabled():
                process_task(STAGE_EMBED, embed_id)
            else:
                ingest_store.update_job_status(task.job_id, IngestJobStatus.EMBEDDING)
        elif stage == STAGE_EMBED:
            ingest_store.update_job_status(task.job_id, IngestJobStatus.EMBEDDING)
            result = run_embed(task)
            ingest_store.finish_task(task_id)
            metrics.inc("tasks.embed.done")
            ingest_store.merge_job_stats(
                task.job_id,
                {"chunks_embedded": int(result.get("embedded", 0))},
            )
            finalize_id = _maybe_enqueue_finalize(task.job_id, async_mode=async_enabled())
            if finalize_id and not async_enabled():
                process_task(STAGE_FINALIZE, finalize_id)
        elif stage == STAGE_FINALIZE:
            run_finalize(task.job_id)
            ingest_store.finish_task(task_id)
            failed = ingest_store.count_tasks(task.job_id, status=IngestTaskStatus.FAILED)
            dead = ingest_store.count_tasks(task.job_id, status=IngestTaskStatus.DEAD)
            terminal = IngestJobStatus.PARTIAL if (failed or dead) else IngestJobStatus.SUCCEEDED
            ingest_store.update_job_status(task.job_id, terminal, mark_finished=True)
            metrics.inc("jobs.succeeded" if terminal == IngestJobStatus.SUCCEEDED else "jobs.partial")
        else:
            raise ValueError(f"unknown stage {stage}")
    except Exception as exc:
        metrics.inc(f"tasks.{stage}.failed")
        _handle_task_failure(task, str(exc))


def _handle_task_failure(task: ingest_store.IngestTask, error: str) -> None:
    if task.attempts < task.max_attempts:
        ingest_store.finish_task(task.id, status=IngestTaskStatus.FAILED, error_msg=error)
        ingest_store.reset_failed_task(task.id)
        if async_enabled():
            ingest_queue.enqueue(
                task.stage,
                {"task_id": task.id, "job_id": task.job_id, "file_key": task.file_key, "retry": True},
            )
        return
    ingest_store.finish_task(task.id, status=IngestTaskStatus.DEAD, error_msg=error)
    if async_enabled():
        ingest_queue.move_to_dlq(task.stage, {"task_id": task.id, "job_id": task.job_id}, error=error)
    ingest_store.update_job_status(task.job_id, IngestJobStatus.PARTIAL, error_msg=error, mark_finished=True)
    metrics.inc("jobs.failed")
