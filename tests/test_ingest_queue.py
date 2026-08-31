"""Tests for ingest Redis queue helpers."""

from __future__ import annotations

import json
from unittest.mock import MagicMock, patch

from rag.ingest import queue as ingest_queue


def test_queue_key_unknown_stage():
    try:
        ingest_queue.queue_key("unknown")
    except ValueError as exc:
        assert "unknown" in str(exc)
    else:
        raise AssertionError("expected ValueError")


@patch("rag.ingest.queue.redis_client")
def test_enqueue_dequeue_roundtrip(mock_client_factory):
    client = MagicMock()
    mock_client_factory.return_value = client
    client.lpush = MagicMock()
    client.brpop = MagicMock(return_value=(ingest_queue.queue_key("parse"), json.dumps({"task_id": 7})))

    ingest_queue.enqueue("parse", {"task_id": 7})
    client.lpush.assert_called_once()

    payload = ingest_queue.dequeue("parse", timeout_sec=1)
    assert payload == {"task_id": 7}
