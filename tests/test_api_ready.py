"""Tests for Python API readiness and internal auth."""

import os
import sys

import pytest

_root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
sys.path.insert(0, _root)

from api.http.app import app  # noqa: E402
from api.auth import (  # noqa: E402
    MIN_SECRET_LEN,
    admin_secret_ok,
    extract_bearer,
    rag_service_token_ok,
    rag_token_from_metadata,
    reload_secrets,
    resolve_rag_token,
    validate_production_secrets,
)
from api.retrieve_metrics import reset_retrieve_metrics_for_tests  # noqa: E402


def _set_rag_token(value: str | None) -> None:
    if value is None:
        os.environ.pop("RAG_SERVICE_TOKEN", None)
    else:
        os.environ["RAG_SERVICE_TOKEN"] = value
    reload_secrets()


def _set_admin_secret(value: str | None) -> None:
    if value is None:
        os.environ.pop("ADMIN_SECRET", None)
    else:
        os.environ["ADMIN_SECRET"] = value
    reload_secrets()


def test_health_public():
    client = app.test_client()
    resp = client.get("/health")
    assert resp.status_code == 200
    assert resp.get_json()["status"] == "healthy"


def test_ready_without_token_when_unconfigured():
    _set_rag_token(None)
    client = app.test_client()
    resp = client.get("/ready")
    assert resp.status_code == 200
    body = resp.get_json()
    assert body["status"] == "ready"
    assert body["checks"]["data"] == "ok"
    assert "index" in body["checks"]


def test_ready_rejects_wrong_token_when_configured():
    _set_rag_token("secret-token")
    try:
        client = app.test_client()
        resp = client.get("/ready")
        assert resp.status_code == 403
        resp2 = client.get("/ready", headers={"X-RAG-Service-Token": "secret-token"})
        assert resp2.status_code == 200
        resp3 = client.get("/ready", headers={"Authorization": "Bearer secret-token"})
        assert resp3.status_code == 200
    finally:
        _set_rag_token(None)


def test_domains_requires_token_when_configured():
    _set_rag_token("secret-token")
    try:
        client = app.test_client()
        assert client.get("/domains").status_code == 403
        ok = client.get("/domains", headers={"X-RAG-Service-Token": "secret-token"})
        assert ok.status_code == 200
        assert ok.get_json()["success"] is True
    finally:
        _set_rag_token(None)


def test_auth_helpers_constant_time_match():
    _set_rag_token(None)
    assert rag_service_token_ok(None) is True
    _set_rag_token("abc")
    try:
        assert rag_service_token_ok("abc") is True
        assert rag_service_token_ok("abd") is False
        assert rag_service_token_ok("") is False
    finally:
        _set_rag_token(None)

    _set_admin_secret(None)
    assert admin_secret_ok("x") is False
    _set_admin_secret("adm")
    try:
        assert admin_secret_ok("adm") is True
        assert admin_secret_ok("bad") is False
    finally:
        _set_admin_secret(None)


def test_resolve_prefers_header_over_bearer():
    assert resolve_rag_token(header_token="from-header", authorization="Bearer from-bearer") == "from-header"
    assert resolve_rag_token(header_token="", authorization="Bearer only-bearer") == "only-bearer"
    assert extract_bearer("bearer xyz") == "xyz"
    assert extract_bearer("Basic abc") == ""


def test_rag_token_from_metadata():
    assert (
        rag_token_from_metadata(
            [("x-rag-service-token", "h"), ("authorization", "Bearer b")]
        )
        == "h"
    )
    assert rag_token_from_metadata([("authorization", "Bearer only")]) == "only"


def test_validate_production_secrets_min_length():
    _set_rag_token("short")
    _set_admin_secret("also-short")
    with pytest.raises(RuntimeError, match="at least"):
        validate_production_secrets(min_len=MIN_SECRET_LEN)
    long = "x" * MIN_SECRET_LEN
    _set_rag_token(long)
    _set_admin_secret(long)
    validate_production_secrets(min_len=MIN_SECRET_LEN)
    _set_rag_token(None)
    _set_admin_secret(None)


def test_metrics_includes_retrieve_counters():
    reset_retrieve_metrics_for_tests()
    client = app.test_client()
    resp = client.get("/metrics")
    assert resp.status_code == 200
    text = resp.get_data(as_text=True)
    assert "rag_retrieve_requests_total" in text
    assert 'protocol="http"' in text
    assert "rag_retrieve_duration_seconds_bucket" in text
    assert "rag_embedding_cache_hit_total" in text


def test_admin_reindex_forbidden_without_secret():
    _set_admin_secret(None)
    client = app.test_client()
    resp = client.post("/admin/reindex")
    assert resp.status_code == 403
