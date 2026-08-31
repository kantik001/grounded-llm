"""Redis list queues for ingest workers."""

from __future__ import annotations

import json
import os
import time
from typing import Any

from rag.ingest.models import STAGE_EMBED, STAGE_FINALIZE, STAGE_PARSE

_QUEUE_PREFIX = "ingest:queue"
_DLQ_KEY = "ingest:dlq"

_STAGE_KEYS = {
    STAGE_PARSE: f"{_QUEUE_PREFIX}:parse",
    STAGE_EMBED: f"{_QUEUE_PREFIX}:embed",
    STAGE_FINALIZE: f"{_QUEUE_PREFIX}:finalize",
}


def _redis_url() -> str:
    return (os.environ.get("REDIS_URL") or "redis://127.0.0.1:6379/0").strip()


def redis_client():
    import redis

    return redis.Redis.from_url(_redis_url(), decode_responses=True)


def queue_key(stage: str) -> str:
    key = _STAGE_KEYS.get(stage)
    if key is None:
        raise ValueError(f"unknown ingest stage: {stage}")
    return key


def enqueue(stage: str, payload: dict[str, Any]) -> None:
    client = redis_client()
    client.lpush(queue_key(stage), json.dumps(payload, separators=(",", ":")))


def dequeue(stage: str, *, timeout_sec: int = 5) -> dict[str, Any] | None:
    client = redis_client()
    item = client.brpop(queue_key(stage), timeout=timeout_sec)
    if not item:
        return None
    _, raw = item
    data = json.loads(raw)
    if not isinstance(data, dict):
        raise ValueError("invalid queue payload")
    return data


def move_to_dlq(stage: str, payload: dict[str, Any], *, error: str) -> None:
    client = redis_client()
    record = {
        "stage": stage,
        "error": error,
        "payload": payload,
        "ts": int(time.time()),
    }
    client.lpush(_DLQ_KEY, json.dumps(record, separators=(",", ":")))


def dlq_depth() -> int:
    client = redis_client()
    return int(client.llen(_DLQ_KEY))
