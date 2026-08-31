"""Ingest job/task status constants."""

from __future__ import annotations

from enum import StrEnum


class IngestJobStatus(StrEnum):
    QUEUED = "queued"
    PARSING = "parsing"
    EMBEDDING = "embedding"
    INDEXING = "indexing"
    SUCCEEDED = "succeeded"
    FAILED = "failed"
    PARTIAL = "partial"


class IngestTaskStatus(StrEnum):
    PENDING = "pending"
    PROCESSING = "processing"
    DONE = "done"
    FAILED = "failed"
    DEAD = "dead"


STAGE_PARSE = "parse"
STAGE_EMBED = "embed"
STAGE_FINALIZE = "finalize"

TERMINAL_JOB_STATUSES = frozenset(
    {IngestJobStatus.SUCCEEDED, IngestJobStatus.FAILED, IngestJobStatus.PARTIAL}
)
