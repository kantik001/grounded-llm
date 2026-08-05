"""Unit tests for api.retrieve_metrics."""

from __future__ import annotations

import os
import sys

_root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
sys.path.insert(0, _root)

from api.retrieve_metrics import (  # noqa: E402
    record_retrieve,
    reset_retrieve_metrics_for_tests,
    retrieve_metrics_lines,
)


def test_record_retrieve_labels_and_histogram():
    reset_retrieve_metrics_for_tests()
    record_retrieve(0.02, protocol="http", ok=True)
    record_retrieve(0.2, protocol="http", ok=False, business_failure=True)
    record_retrieve(1.5, protocol="grpc", ok=False, business_failure=False)

    text = "\n".join(retrieve_metrics_lines())
    assert 'rag_retrieve_requests_total{protocol="http",outcome="ok"} 1' in text
    assert 'rag_retrieve_requests_total{protocol="http",outcome="business"} 1' in text
    assert 'rag_retrieve_requests_total{protocol="grpc",outcome="error"} 1' in text
    assert 'rag_retrieve_duration_seconds_bucket{protocol="http",le="0.025"}' in text
    assert 'rag_retrieve_duration_seconds_bucket{protocol="grpc",le="+Inf"}' in text
    assert 'rag_retrieve_duration_seconds_sum{protocol="http"}' in text
    assert 'rag_retrieve_duration_seconds_count{protocol="grpc"} 1' in text


def test_metrics_endpoint_exposes_histogram_labels():
    from api.http.app import app

    reset_retrieve_metrics_for_tests()
    record_retrieve(0.01, protocol="http", ok=True)
    client = app.test_client()
    text = client.get("/metrics").get_data(as_text=True)
    assert 'outcome="ok"' in text
    assert "rag_retrieve_duration_seconds_bucket" in text
    assert "rag_embedding_cache_hit_total" in text
